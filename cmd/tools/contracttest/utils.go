package contracttest

import (
	"math/rand"
	"testing"

	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"math/big"

	"fmt"

	"time"

	"github.com/SmartMeshFoundation/SmartRaiden/network/rpc/contracts"
	"github.com/SmartMeshFoundation/SmartRaiden/transfer/mtree"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// CoOperativeSettleForContracts : param for CoOperativeSettle
type CoOperativeSettleForContracts struct {
	Particiant1         common.Address
	Participant2        common.Address
	Participant1Balance *big.Int
	Participant2Balance *big.Int
	ChannelIdentifier   contracts.ChannelIdentifier
	OpenBlockNumber     uint64
	TokenNetworkAddress common.Address
	ChainID             *big.Int
}

func (c *CoOperativeSettleForContracts) sign(key *ecdsa.PrivateKey) []byte {
	buf := new(bytes.Buffer)
	_, err := buf.Write(c.Particiant1[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(c.Participant1Balance))
	_, err = buf.Write(c.Participant2[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(c.Participant2Balance))
	_, err = buf.Write(c.ChannelIdentifier[:])
	err = binary.Write(buf, binary.BigEndian, c.OpenBlockNumber)
	_, err = buf.Write(utils.BigIntTo32Bytes(c.ChainID))
	sig, err := utils.SignData(key, buf.Bytes())
	if err != nil {
		panic(err)
	}
	return sig
}

// WithDraw1ForContract : param for withdraw 1
type WithDraw1ForContract struct {
	Participant1         common.Address
	Participant2         common.Address
	Participant1Deposit  *big.Int
	Participant2Deposit  *big.Int
	Participant1Withdraw *big.Int
	ChannelIdentifier    contracts.ChannelIdentifier
	OpenBlockNumber      uint64
	TokenNetworkAddress  common.Address
	ChainID              *big.Int
}

func (w *WithDraw1ForContract) sign(key *ecdsa.PrivateKey) []byte {
	buf := new(bytes.Buffer)
	_, err := buf.Write(w.Participant1[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(w.Participant1Deposit))
	_, err = buf.Write(w.Participant2[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(w.Participant2Deposit))
	_, err = buf.Write(utils.BigIntTo32Bytes(w.Participant1Withdraw))
	_, err = buf.Write(w.ChannelIdentifier[:])
	err = binary.Write(buf, binary.BigEndian, w.OpenBlockNumber)
	//buf.Write(w.TokenNetworkAddress[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(w.ChainID))
	sig, err := utils.SignData(key, buf.Bytes())
	if err != nil {
		panic(err)
	}
	return sig
}

// WithDraw2ForContract : param for withdraw 2
type WithDraw2ForContract struct {
	Participant1         common.Address
	Participant2         common.Address
	Participant1Deposit  *big.Int
	Participant2Deposit  *big.Int
	Participant1Withdraw *big.Int
	Participant2Withdraw *big.Int
	ChannelIdentifier    contracts.ChannelIdentifier
	OpenBlockNumber      uint64
	TokenNetworkAddress  common.Address
	ChainID              *big.Int
}

func (w *WithDraw2ForContract) sign(key *ecdsa.PrivateKey) []byte {
	buf := new(bytes.Buffer)
	_, err := buf.Write(w.Participant1[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(w.Participant1Deposit))
	_, err = buf.Write(w.Participant2[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(w.Participant2Deposit))
	_, err = buf.Write(utils.BigIntTo32Bytes(w.Participant1Withdraw))
	_, err = buf.Write(utils.BigIntTo32Bytes(w.Participant2Withdraw))
	_, err = buf.Write(w.ChannelIdentifier[:])
	err = binary.Write(buf, binary.BigEndian, w.OpenBlockNumber)
	//buf.Write(w.TokenNetworkAddress[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(w.ChainID))
	sig, err := utils.SignData(key, buf.Bytes())
	if err != nil {
		panic(err)
	}
	return sig
}

func cooperativeSettleChannelIfExists(a1 *Account, a2 *Account) {
	cs := getCooperativeSettleParams(a1, a2, big.NewInt(0), big.NewInt(0))
	if cs == nil {
		return
	}
	tx, err := env.TokenNetwork.CooperativeSettle(
		a1.Auth,
		a1.Address, cs.Participant1Balance,
		a2.Address, cs.Participant2Balance,
		cs.sign(a1.Key),
		cs.sign(a2.Key))
	if err != nil {
		panic(err)
	}
	_, err = bind.WaitMined(context.Background(), env.Client, tx)
	if err != nil {
		panic(err)
	}
}

func (env *Env) getTwoRandomAccount(t *testing.T) (*Account, *Account) {
	var index1, index2 int
	n := len(env.Accounts)
	seed := rand.NewSource(time.Now().Unix())
	r1 := rand.New(seed)
	index1 = r1.Intn(n)
	index2 = r1.Intn(n)
	for index1 == index2 {
		index2 = r1.Intn(n)
	}
	return env.Accounts[index1], env.Accounts[index2]
}

func (env *Env) getTwoAccountWithoutChannelClose(t *testing.T) (*Account, *Account) {
	for index1, a1 := range env.Accounts {
		for index2, a2 := range env.Accounts {
			if index1 == index2 {
				continue
			}
			_, _, _, channelState, _, _ := getChannelInfo(a1, a2)
			if channelState != ChannelStateClosed {
				return a1, a2
			}
		}
	}
	panic("no usable account, need to run cmd/newTestEnv")
}

func (env *Env) getThreeRandomAccount(t *testing.T) (*Account, *Account, *Account) {
	var index1, index2, index3 int
	n := len(env.Accounts)
	seed := rand.NewSource(time.Now().Unix())
	r1 := rand.New(seed)
	index1 = r1.Intn(n)
	index2 = r1.Intn(n)
	index3 = r1.Intn(n)
	for index1 == index2 {
		index2 = r1.Intn(n)
	}
	for index3 == index1 || index3 == index2 {
		index3 = r1.Intn(n)
	}
	return env.Accounts[index1], env.Accounts[index2], env.Accounts[index3]
}

func (env *Env) getRandomAccountExcept(t *testing.T, accounts ...*Account) *Account {
	usable := true
	for _, account := range env.Accounts {
		usable = true
		for _, t := range accounts {
			if account.Address.String() == t.Address.String() {
				usable = false
				break
			}
		}
		if usable {
			return account
		}
	}
	panic("no usable account")
}

func getCooperativeSettleParams(a1, a2 *Account, balanceA1, balanceA2 *big.Int) *CoOperativeSettleForContracts {
	var err error
	channelID, _, openBlockNumber, state, _, ChainID := getChannelInfo(a1, a2)
	if state != ChannelStateOpened {
		return nil
	}
	if state == ChannelStateOpened {
		balanceA1, _, _, err = env.TokenNetwork.GetChannelParticipantInfo(nil, a1.Address, a2.Address)
		if err != nil {
			panic(err)
		}
		balanceA2, _, _, err = env.TokenNetwork.GetChannelParticipantInfo(nil, a2.Address, a1.Address)
		if err != nil {
			panic(err)
		}
	}
	return &CoOperativeSettleForContracts{
		Particiant1:         a1.Address,
		Participant2:        a2.Address,
		Participant1Balance: balanceA1,
		Participant2Balance: balanceA2,
		ChannelIdentifier:   channelID,
		OpenBlockNumber:     openBlockNumber,
		ChainID:             ChainID,
		TokenNetworkAddress: env.TokenNetworkAddress,
	}
}

func openChannelAndDeposit(a1, a2 *Account, depositA1, depositA2 *big.Int, settleTimeout uint64) {
	cooperativeSettleChannelIfExists(a1, a2)
	var tx *types.Transaction
	var err error
	if depositA1.Int64() > 0 {
		tx, err = env.TokenNetwork.OpenChannelWithDeposit(a1.Auth, a1.Address, a2.Address, settleTimeout, depositA1)
	} else {
		tx, err = env.TokenNetwork.OpenChannel(a1.Auth, a1.Address, a2.Address, settleTimeout)
	}
	if err != nil {
		panic(err)
	}
	_, err = bind.WaitMined(context.Background(), env.Client, tx)
	if err != nil {
		panic(err)
	}
	if depositA2.Int64() > 0 {
		tx, err = env.TokenNetwork.Deposit(a2.Auth, a2.Address, a1.Address, depositA2)
		if err != nil {
			panic(err)
		}
		_, err = bind.WaitMined(context.Background(), env.Client, tx)
		if err != nil {
			panic(err)
		}
	}
}

func withdraw(a1 *Account, depositA1, withdrawA1 *big.Int, a2 *Account, depositA2, withdrawA2 *big.Int) {
	channelID, _, openBlockNumber, _, _, ChainID := getChannelInfo(a1, a2)
	param1 := &WithDraw1ForContract{
		Participant1:         a1.Address,
		Participant2:         a2.Address,
		Participant1Deposit:  depositA1,
		Participant2Deposit:  depositA2,
		Participant1Withdraw: withdrawA1,
		ChannelIdentifier:    channelID,
		OpenBlockNumber:      openBlockNumber,
		TokenNetworkAddress:  env.TokenNetworkAddress,
		ChainID:              ChainID,
	}
	param2 := &WithDraw2ForContract{
		Participant1:         a1.Address,
		Participant2:         a2.Address,
		Participant1Deposit:  depositA1,
		Participant2Deposit:  depositA2,
		Participant1Withdraw: withdrawA1,
		Participant2Withdraw: withdrawA2,
		ChannelIdentifier:    channelID,
		OpenBlockNumber:      openBlockNumber,
		TokenNetworkAddress:  env.TokenNetworkAddress,
		ChainID:              ChainID,
	}
	tx, err := env.TokenNetwork.WithDraw(
		a2.Auth,
		param2.Participant1,
		param2.Participant1Deposit,
		param2.Participant1Withdraw,
		param2.Participant2,
		param2.Participant2Deposit,
		param2.Participant2Withdraw,
		param1.sign(a1.Key),
		param2.sign(a2.Key),
	)
	if err != nil {
		panic(err)
	}
	_, err = bind.WaitMined(context.Background(), env.Client, tx)
	if err != nil {
		panic(err)
	}
}

func createWithdrawParam(a1 *Account, depositA1, withdrawA1 *big.Int, a2 *Account, depositA2, withdrawA2 *big.Int) (*WithDraw1ForContract, *WithDraw2ForContract) {
	channelID, _, openBlockNumber, _, _, ChainID := getChannelInfo(a1, a2)
	param1 := &WithDraw1ForContract{
		Participant1:         a1.Address,
		Participant2:         a2.Address,
		Participant1Deposit:  depositA1,
		Participant2Deposit:  depositA2,
		Participant1Withdraw: withdrawA1,
		ChannelIdentifier:    channelID,
		OpenBlockNumber:      openBlockNumber,
		TokenNetworkAddress:  env.TokenNetworkAddress,
		ChainID:              ChainID,
	}
	param2 := &WithDraw2ForContract{
		Participant1:         a1.Address,
		Participant2:         a2.Address,
		Participant1Deposit:  depositA1,
		Participant2Deposit:  depositA2,
		Participant1Withdraw: withdrawA1,
		Participant2Withdraw: withdrawA2,
		ChannelIdentifier:    channelID,
		OpenBlockNumber:      openBlockNumber,
		TokenNetworkAddress:  env.TokenNetworkAddress,
		ChainID:              ChainID,
	}
	return param1, param2
}

//BalanceData of contract
type BalanceData struct {
	TransferAmount *big.Int
	LocksRoot      common.Hash
}

// Hash :
func (b *BalanceData) Hash() common.Hash {
	buf := new(bytes.Buffer)
	_, err := buf.Write(b.LocksRoot[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(b.TransferAmount))
	if err != nil {
		panic(err)
	}
	return utils.Sha3(buf.Bytes())
}

//BalanceProofForContract for contract
type BalanceProofForContract struct {
	AdditionalHash      common.Hash
	ChannelIdentifier   contracts.ChannelIdentifier
	TokenNetworkAddress common.Address
	ChainID             *big.Int
	Signature           []byte
	OpenBlockNumber     uint64
	Nonce               uint64
	BalanceData
}

func (b *BalanceProofForContract) sign(key *ecdsa.PrivateKey) {
	buf := new(bytes.Buffer)
	_, err := buf.Write(utils.BigIntTo32Bytes(b.TransferAmount))
	_, err = buf.Write(b.LocksRoot[:])
	err = binary.Write(buf, binary.BigEndian, b.Nonce)
	_, err = buf.Write(b.AdditionalHash[:])
	_, err = buf.Write(b.ChannelIdentifier[:])
	err = binary.Write(buf, binary.BigEndian, b.OpenBlockNumber)
	//buf.Write(b.TokenNetworkAddress[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(b.ChainID))
	sig, err := utils.SignData(key, buf.Bytes())
	if err != nil {
		panic(err)
	}
	b.Signature = sig
}

func (b *BalanceProofForContract) signWithoutChange(key *ecdsa.PrivateKey) []byte {
	buf := new(bytes.Buffer)
	_, err := buf.Write(utils.BigIntTo32Bytes(b.TransferAmount))
	_, err = buf.Write(b.LocksRoot[:])
	err = binary.Write(buf, binary.BigEndian, b.Nonce)
	_, err = buf.Write(b.AdditionalHash[:])
	_, err = buf.Write(b.ChannelIdentifier[:])
	err = binary.Write(buf, binary.BigEndian, b.OpenBlockNumber)
	//buf.Write(b.TokenNetworkAddress[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(b.ChainID))
	sig, err := utils.SignData(key, buf.Bytes())
	if err != nil {
		panic(err)
	}
	return sig
}

// BalanceProofUpdateForContracts :
type BalanceProofUpdateForContracts struct {
	BalanceProofForContract
	NonClosingSignature []byte
}

func (b *BalanceProofUpdateForContracts) sign(key *ecdsa.PrivateKey) {
	buf := new(bytes.Buffer)
	_, err := buf.Write(utils.BigIntTo32Bytes(b.TransferAmount))
	_, err = buf.Write(b.LocksRoot[:])
	err = binary.Write(buf, binary.BigEndian, b.Nonce)
	_, err = buf.Write(b.AdditionalHash[:])
	_, err = buf.Write(b.ChannelIdentifier[:])
	err = binary.Write(buf, binary.BigEndian, b.OpenBlockNumber)
	//buf.Write(b.TokenNetworkAddress[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(b.ChainID))
	_, err = buf.Write(b.Signature)
	sig, err := utils.SignData(key, buf.Bytes())
	if err != nil {
		panic(err)
	}
	b.NonClosingSignature = sig
}

func createPartnerBalanceProof(self *Account, partner *Account, transferAmount *big.Int, locksroot common.Hash, additionalHash common.Hash, nonce uint64) *BalanceProofForContract {
	channelID, _, openBlockNumber, _, _, ChainID := getChannelInfo(self, partner)
	bd := &BalanceData{
		TransferAmount: transferAmount,
		LocksRoot:      locksroot,
	}
	bp := &BalanceProofForContract{
		BalanceData:         *bd,
		OpenBlockNumber:     openBlockNumber,
		AdditionalHash:      additionalHash,
		ChannelIdentifier:   channelID,
		TokenNetworkAddress: env.TokenNetworkAddress,
		ChainID:             ChainID,
		Nonce:               nonce,
	}
	bp.sign(partner.Key)
	return bp
}

func createLock(expiredBlock int64, amounts ...*big.Int) (locks []*mtree.Lock, secrets []common.Hash) {
	for i := 0; i < len(amounts); i++ {
		secret := utils.Sha3([]byte(utils.RandomString(10)))
		secrets = append(secrets, secret)
		l := &mtree.Lock{
			Expiration:     expiredBlock,
			Amount:         amounts[i],
			LockSecretHash: utils.Sha3(secret[:]),
		}
		locks = append(locks, l)
	}
	return
}

func createLockByArray(expiredBlock int64, amounts []*big.Int) (locks []*mtree.Lock, secrets []common.Hash) {
	for i := 0; i < len(amounts); i++ {
		secret := utils.Sha3([]byte(utils.RandomString(10)))
		secrets = append(secrets, secret)
		l := &mtree.Lock{
			Expiration:     expiredBlock,
			Amount:         amounts[i],
			LockSecretHash: utils.Sha3(secret[:]),
		}
		locks = append(locks, l)
	}
	return
}

func registrySecrets(account *Account, secrets []common.Hash) {
	maxLocks := 5
	for i := 0; i < len(secrets); i++ {
		s := secrets[i]
		if i > maxLocks { //最多注册前五个密码,否则太浪费时间了.
			break
		}
		tx, err := env.SecretRegistry.RegisterSecret(account.Auth, s)
		if err != nil {
			panic(err)
		}
		r, err := bind.WaitMined(context.Background(), env.Client, tx)
		if err != nil {
			panic(err)
		}
		if r.Status != types.ReceiptStatusSuccessful {
			panic(err)
		}
	}
}

func getChannelInfo(a1 *Account, a2 *Account) (channelID [32]byte, settleBlockNum uint64, openBlockNumber uint64, state uint8, settleTimeout uint64, ChainID *big.Int) {
	channelID, settleBlockNum, openBlockNumber, state, settleTimeout, err := env.TokenNetwork.GetChannelInfo(nil, a1.Address, a2.Address)
	if err != nil {
		panic(err)
	}
	ChainID, err = env.TokenNetwork.ChainId(nil)
	if err != nil {
		panic(err)
	}
	return
}

func approve(account *Account, amount *big.Int) {
	tx, err := env.Token.Approve(account.Auth, env.TokenNetworkAddress, amount)
	if err != nil {
		panic(err)
	}
	r, err := bind.WaitMined(context.Background(), env.Client, tx)
	if err != nil {
		panic(err)
	}
	if r.Status != types.ReceiptStatusSuccessful {
		panic(err)
	}
}

func endMsg(name string, count int, accounts ...*Account) string {
	msg := fmt.Sprintf("%s完成 CaseNum=%d", name, count)
	if accounts != nil && len(accounts) > 0 {
		msg = msg + " 使用账号:"
	}
	for index, account := range accounts {
		msg = msg + fmt.Sprintf("a%d=%s ", index+1, account.Address.String())
	}
	return msg
}

func getTokenBalance(account *Account) *big.Int {
	balance, err := env.Token.BalanceOf(nil, account.Address)
	if err != nil {
		panic(err)
	}
	return balance
}

func getTokenBalanceByAddess(address common.Address) *big.Int {
	balance, err := env.Token.BalanceOf(nil, address)
	if err != nil {
		panic(err)
	}
	return balance
}

func waitToSettle(a1 *Account, a2 *Account) {
	_, settleBlockNum, _, _, _, _ := getChannelInfo(a1, a2)
	punishBlockNumber, err := env.TokenNetwork.PunishBlockNumber(nil)
	if err != nil {
		panic(err)
	}
	waitUntilBlockNo(settleBlockNum + punishBlockNumber)
}

func waitUntilBlockNo(blockNo uint64) {
	fmt.Printf("wait until block %d ...\n", blockNo)
	for {
		var h *types.Header
		h, err := env.Client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			panic(err)
		}
		if h.Number.Uint64() >= blockNo {
			break
		}
		time.Sleep(time.Second)
	}
}

func waitByBlocknum(blocknum uint64) {
	h, err := env.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	waitUntilBlockNo(h.Number.Uint64() + blocknum)
}

func waitToPunish(a1, a2 *Account) {
	_, settleBlockNum, _, _, _, _ := getChannelInfo(a1, a2)
	waitUntilBlockNo(settleBlockNum + 1)
}

func waitToUpdateBalanceProofDelegate(a1, a2 *Account) {
	_, settleBlockNum, _, _, settleTimeout, _ := getChannelInfo(a1, a2)
	waitUntilBlockNo(settleBlockNum - settleTimeout/2)
}

func getLatestBlockNumber() *types.Header {
	h, err := env.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	return h
}

// UnlockDelegateForContract :
type UnlockDelegateForContract struct {
	Agent             common.Address
	Expiraition       int64
	Amount            *big.Int
	SecretHash        common.Hash
	ChannelIdentifier contracts.ChannelIdentifier
	ChainID           *big.Int
	OpenBlockNumber   uint64
	MerkleProof       []byte
}

func (u *UnlockDelegateForContract) sign(key *ecdsa.PrivateKey) []byte {
	buf := new(bytes.Buffer)
	_, err := buf.Write(u.Agent[:])
	_, err = buf.Write(utils.BigIntTo32Bytes(big.NewInt(u.Expiraition)))
	_, err = buf.Write(utils.BigIntTo32Bytes(u.Amount))
	_, err = buf.Write(u.SecretHash[:])
	_, err = buf.Write(u.ChannelIdentifier[:])
	err = binary.Write(buf, binary.BigEndian, u.OpenBlockNumber)
	_, err = buf.Write(utils.BigIntTo32Bytes(u.ChainID))
	sig, err := utils.SignData(key, buf.Bytes())
	if err != nil {
		panic(err)
	}
	return sig
}
