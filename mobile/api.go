package mobile

import (
	"encoding/json"

	"fmt"
	"time"

	"math/big"

	"errors"

	"github.com/SmartMeshFoundation/SmartRaiden"
	"github.com/SmartMeshFoundation/SmartRaiden/internal/rpanic"
	"github.com/SmartMeshFoundation/SmartRaiden/log"
	"github.com/SmartMeshFoundation/SmartRaiden/network"
	"github.com/SmartMeshFoundation/SmartRaiden/network/netshare"
	"github.com/SmartMeshFoundation/SmartRaiden/params"
	"github.com/SmartMeshFoundation/SmartRaiden/restful/v1"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"github.com/ethereum/go-ethereum/common"
)

//API for export interface
type API struct {
	api *smartraiden.RaidenAPI
}

func marshal(v interface{}) (s string, err error) {
	d, err := json.Marshal(v)
	if err != nil {
		log.Error(err.Error())
		return
	}
	return string(d), nil
}

//GetChannelList GET /api/1/channels
func (a *API) GetChannelList() (channels string, err error) {
	defer func() {
		log.Trace(fmt.Sprintf("ApiCall GetChannelList channels=\n%s,err=%v", channels, err))
	}()
	chs, err := a.api.GetChannelList(utils.EmptyAddress, utils.EmptyAddress)
	if err != nil {
		log.Error(err.Error())
		return
	}
	var datas []*v1.ChannelData
	for _, c := range chs {
		d := &v1.ChannelData{
			ChannelAddress:      common.BytesToHash(c.Key).String(),
			PartnerAddrses:      c.PartnerAddress().String(),
			Balance:             c.OurBalance(),
			PartnerBalance:      c.PartnerBalance(),
			LockedAmount:        c.OurAmountLocked(),
			PartnerLockedAmount: c.PartnerAmountLocked(),
			State:               c.State,
			TokenAddress:        c.TokenAddress().String(),
			SettleTimeout:       c.SettleTimeout,
			RevealTimeout:       c.RevealTimeout,
		}
		datas = append(datas, d)
	}
	channels, err = marshal(datas)
	return
}

//GetOneChannel GET /api/1/channels/0x2a65aca4d5fc5b5c859090a6c34d164135398226
func (a *API) GetOneChannel(channelAddress string) (channel string, err error) {
	defer func() {
		log.Trace(fmt.Sprintf("Api GetOneChannel in channel address=%s,out channel=\n%s,err=%v", channelAddress, channel, err))
	}()
	chaddr := common.HexToHash(channelAddress)
	c, err := a.api.GetChannel(chaddr)
	if err != nil {
		log.Error(err.Error())
		return
	}
	d := &v1.ChannelDataDetail{
		ChannelAddress:           common.BytesToHash(c.Key).String(),
		PartnerAddrses:           c.PartnerAddress().String(),
		Balance:                  c.OurBalance(),
		PartnerBalance:           c.PartnerBalance(),
		State:                    c.State,
		SettleTimeout:            c.SettleTimeout,
		TokenAddress:             c.TokenAddress().String(),
		LockedAmount:             c.OurAmountLocked(),
		PartnerLockedAmount:      c.PartnerAmountLocked(),
		ClosedBlock:              c.ClosedBlock,
		SettledBlock:             c.SettledBlock,
		OurLeaves:                c.OurLeaves,
		PartnerLeaves:            c.PartnerLeaves,
		OurKnownSecretLocks:      c.OurLock2UnclaimedLocks(),
		OurUnkownSecretLocks:     c.OurLock2PendingLocks(),
		PartnerUnkownSecretLocks: c.PartnerLock2PendingLocks(),
		PartnerKnownSecretLocks:  c.PartnerLock2UnclaimedLocks(),
		OurBalanceProof:          c.OurBalanceProof,
		PartnerBalanceProof:      c.PartnerBalanceProof,
	}
	channel, err = marshal(d)
	return
}

//OpenChannel put request
func (a *API) OpenChannel(partnerAddress, tokenAddress string, settleTimeout int, balanceStr string) (channel string, err error) {
	defer func() {
		log.Trace(fmt.Sprintf("Api OpenChannel in partnerAddress=%s,tokenAddress=%s,settletTimeout=%d,balanceStr=%s\nout channel=\n%s,err=%v",
			partnerAddress, tokenAddress, settleTimeout, balanceStr, channel, err,
		))
	}()
	partnerAddr := common.HexToAddress(partnerAddress)
	tokenAddr := common.HexToAddress(tokenAddress)
	balance, _ := new(big.Int).SetString(balanceStr, 0)
	c, err := a.api.Open(tokenAddr, partnerAddr, settleTimeout, params.DefaultRevealTimeout, balance)
	if err != nil {
		log.Error(err.Error())
		return
	}
	d := &v1.ChannelData{
		ChannelAddress:      common.BytesToHash(c.Key).String(),
		PartnerAddrses:      c.PartnerAddress().String(),
		Balance:             c.OurBalance(),
		PartnerBalance:      c.PartnerBalance(),
		State:               c.State,
		SettleTimeout:       c.SettleTimeout,
		TokenAddress:        c.TokenAddress().String(),
		LockedAmount:        c.OurAmountLocked(),
		PartnerLockedAmount: c.PartnerAmountLocked(),
	}
	channel, err = marshal(d)
	return

}

//CloseChannel close a channel
func (a *API) CloseChannel(channelAddress string, force bool) (channel string, err error) {
	defer func() {
		log.Trace(fmt.Sprintf("Api CloseChannel in channelAddress=%s,out channel=\n%s,err=%v",
			channelAddress, channel, err,
		))
	}()
	chAddr := common.HexToHash(channelAddress)
	c, err := a.api.GetChannel(chAddr)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if force {
		c, err = a.api.Close(c.TokenAddress(), c.PartnerAddress())
		if err != nil {
			log.Error(err.Error())
			return
		}
	} else {
		c, err = a.api.CooperativeSettle(c.TokenAddress(), c.PartnerAddress())
		if err != nil {
			log.Error(err.Error())
			return
		}
	}
	d := &v1.ChannelData{
		ChannelAddress:      common.BytesToHash(c.Key).String(),
		PartnerAddrses:      c.PartnerAddress().String(),
		Balance:             c.OurBalance(),
		PartnerBalance:      c.PartnerBalance(),
		State:               c.State,
		SettleTimeout:       c.SettleTimeout,
		TokenAddress:        c.TokenAddress().String(),
		LockedAmount:        c.OurAmountLocked(),
		PartnerLockedAmount: c.PartnerAmountLocked(),
	}
	channel, err = marshal(d)
	return
}

//SettleChannel settle a channel
func (a *API) SettleChannel(channelAddres string) (channel string, err error) {
	defer func() {
		log.Trace(fmt.Sprintf("Api SettleChannel in channelAddress=%s,out channel=\n%s,err=%v",
			channelAddres, channel, err,
		))
	}()

	chAddr := common.HexToHash(channelAddres)
	c, err := a.api.GetChannel(chAddr)
	if err != nil {
		log.Error(err.Error())
		return
	}
	c, err = a.api.Settle(c.TokenAddress(), c.PartnerAddress())
	if err != nil {
		log.Error(err.Error())
		return
	}
	d := &v1.ChannelData{
		ChannelAddress:      common.BytesToHash(c.Key).String(),
		PartnerAddrses:      c.PartnerAddress().String(),
		Balance:             c.OurBalance(),
		PartnerBalance:      c.PartnerBalance(),
		State:               c.State,
		SettleTimeout:       c.SettleTimeout,
		TokenAddress:        c.TokenAddress().String(),
		LockedAmount:        c.OurAmountLocked(),
		PartnerLockedAmount: c.PartnerAmountLocked(),
	}
	channel, err = marshal(d)
	return
}

//DepositChannel deposit balance to channel
func (a *API) DepositChannel(channelAddres string, balanceStr string) (channel string, err error) {
	defer func() {
		log.Trace(fmt.Sprintf("Api DepositChannel channelAddres=%s,balanceStr=%s,out channel=\n%s,err=%v",
			channelAddres, balanceStr, channel, err,
		))
	}()
	chAddr := common.HexToHash(channelAddres)
	balance, _ := new(big.Int).SetString(balanceStr, 0)
	c, err := a.api.GetChannel(chAddr)
	if err != nil {
		log.Error(fmt.Sprintf("GetChannel %s err %s", utils.HPex(chAddr), err))
		return
	}
	c, err = a.api.Deposit(c.TokenAddress(), c.PartnerAddress(), balance, params.DefaultPollTimeout)
	if err != nil {
		log.Error(fmt.Sprintf("Deposit to %s:%s err %s", utils.APex(c.TokenAddress()),
			utils.APex(c.PartnerAddress()), err))
		return
	}

	d := &v1.ChannelData{
		ChannelAddress:      common.BytesToHash(c.Key).String(),
		PartnerAddrses:      c.PartnerAddress().String(),
		Balance:             c.OurBalance(),
		PartnerBalance:      c.PartnerBalance(),
		State:               c.State,
		SettleTimeout:       c.SettleTimeout,
		TokenAddress:        c.TokenAddress().String(),
		LockedAmount:        c.OurAmountLocked(),
		PartnerLockedAmount: c.PartnerAmountLocked(),
	}
	channel, err = marshal(d)
	return
}

//NetworkEvent GET /api/<version>/events/network
func (a *API) NetworkEvent(fromBlock, toBlock int64) (eventsString string, err error) {
	events, err := a.api.GetNetworkEvents(fromBlock, toBlock)
	if err != nil {
		log.Error(err.Error())
		return
	}
	eventsString, err = marshal(events)
	return
}

//TokensEvent GET /api/1/events/tokens/0x61c808d82a3ac53231750dadc13c777b59310bd9
func (a *API) TokensEvent(fromBlock, toBlock int64, tokenAddress string) (eventsString string, err error) {
	token := common.HexToAddress(tokenAddress)
	events, err := a.api.GetTokenNetworkEvents(token, fromBlock, toBlock)
	if err != nil {
		log.Error(err.Error())
		return
	}
	eventsString, err = marshal(events)
	return
}

//ChannelsEvent GET /api/1/events/channels/0x2a65aca4d5fc5b5c859090a6c34d164135398226?from_block=1337
func (a *API) ChannelsEvent(fromBlock, toBlock int64, channelAddress string) (eventsString string, err error) {
	channel := common.HexToHash(channelAddress)
	events, err := a.api.GetChannelEvents(channel, fromBlock, toBlock)
	if err != nil {
		log.Error(err.Error())
		return
	}
	eventsString, err = marshal(events)
	return
}

//Address GET /api/1/address
func (a *API) Address() (addr string) {
	return a.api.Address().String()
}

//Tokens GET /api/1/tokens
func (a *API) Tokens() (tokens string) {
	tokens, err := marshal(a.api.Tokens())
	if err != nil {
		log.Error(fmt.Sprintf("marshal tokens error %s", err))
	}
	return
}

type partnersData struct {
	PartnerAddress string `json:"partner_address"`
	Channel        string `json:"channel"`
}

//TokenPartners GET /api/1/tokens/0x61bb630d3b2e8eda0fc1d50f9f958ec02e3969f6/partners
func (a *API) TokenPartners(tokenAddress string) (channels string, err error) {
	tokenAddr := common.HexToAddress(tokenAddress)
	chs, err := a.api.GetChannelList(tokenAddr, utils.EmptyAddress)
	if err != nil {
		log.Error(err.Error())
		return
	}
	var datas []*partnersData
	for _, c := range chs {
		d := &partnersData{
			PartnerAddress: c.PartnerAddress().String(),
			Channel:        "api/1/channles/" + c.OurAddress.String(),
		}
		datas = append(datas, d)
	}
	channels, err = marshal(datas)
	return
}

//RegisterToken PUT /api/1/tokens/0xea674fdde714fd979de3edf0f56aa9716b898ec8 Registering a Token
func (a *API) RegisterToken(tokenAddress string) (managerAddress string, err error) {
	defer func() {
		log.Trace(fmt.Sprintf("Api RegisterToken tokenAddress=%s,managerAddress=%s,err=%v",
			tokenAddress, managerAddress, err,
		))
	}()
	tokenAddr := common.HexToAddress(tokenAddress)
	mgr, err := a.api.RegisterToken(tokenAddr)
	if err != nil {
		log.Error(err.Error())
		return
	}
	return mgr.String(), err
}

/*
Transfers POST /api/1/transfers/0x2a65aca4d5fc5b5c859090a6c34d164135398226/0x61c808d82a3ac53231750dadc13c777b59310bd9
Initiating a Transfer
identifier:0 means random identifier generated by system
*/
func (a *API) Transfers(tokenAddress, targetAddress string, amountstr string, feestr string, lockSecretHashstr string, isDirect bool) (transfer string, err error) {
	defer func() {
		log.Trace(fmt.Sprintf("Api Transfers tokenAddress=%s,targetAddress=%s,amountstr=%s,feestr=%s,id=%s,isDirect=%v,\nout transfer=\n%s,err=%v",
			tokenAddress, targetAddress, amountstr, feestr, lockSecretHashstr, isDirect, transfer, err,
		))
	}()
	tokenAddr := common.HexToAddress(tokenAddress)
	targetAddr := common.HexToAddress(targetAddress)
	amount, _ := new(big.Int).SetString(amountstr, 0)
	fee, _ := new(big.Int).SetString(feestr, 0)
	lockSecretHash := common.HexToHash(lockSecretHashstr)
	if amount.Cmp(utils.BigInt0) <= 0 {
		err = errors.New("amount should be positive")
		return
	}
	err = a.api.Transfer(tokenAddr, amount, fee, targetAddr, lockSecretHash, params.MaxRequestTimeout, isDirect)
	if err != nil {
		log.Error(err.Error())
		return
	}
	req := &v1.TransferData{}
	req.Initiator = a.api.Raiden.NodeAddress.String()
	req.Target = targetAddress
	req.Token = tokenAddress
	req.Amount = amount
	req.LockSecretHash = lockSecretHashstr
	req.Fee = fee
	return marshal(req)
}

/*
TokenSwap token swap for maker
role: "maker" or "taker"
*/
func (a *API) TokenSwap(role string, Identifier string, SendingAmountStr, ReceivingAmountStr string, SendingToken, ReceivingToken, TargetAddress string) (err error) {
	type Req struct {
		Role            string   `json:"role"`
		SendingAmount   *big.Int `json:"sending_amount"`
		SendingToken    string   `json:"sending_token"`
		ReceivingAmount int64    `json:"receiving_amount"`
		ReceivingToken  *big.Int `json:"receiving_token"`
	}

	var target common.Address
	target = common.HexToAddress(TargetAddress)
	if len(Identifier) <= 0 {
		err = errors.New("LockSecretHash must not be empty")
		return
	}
	SendingAmount, _ := new(big.Int).SetString(SendingAmountStr, 0)
	ReceivingAmount, _ := new(big.Int).SetString(ReceivingAmountStr, 0)
	if role == "maker" {
		err = a.api.TokenSwapAndWait(Identifier, common.HexToAddress(SendingToken), common.HexToAddress(ReceivingToken),
			a.api.Raiden.NodeAddress, target, SendingAmount, ReceivingAmount)
	} else if role == "taker" {
		err = a.api.ExpectTokenSwap(Identifier, common.HexToAddress(ReceivingToken), common.HexToAddress(SendingToken),
			target, a.api.Raiden.NodeAddress, ReceivingAmount, SendingAmount)
	} else {
		err = fmt.Errorf("provided invalid token swap role %s", role)
	}
	return
}

//Stop stop raiden
func (a *API) Stop() {
	log.Trace("Api Stop")
	//test only
	a.api.Stop()
}

/*
ChannelFor3rdParty generate info for 3rd party use,
for update transfer and withdraw.
*/
func (a *API) ChannelFor3rdParty(channelAddress, thirdPartyAddress string) (r string, err error) {
	channelAddr := common.HexToHash(channelAddress)
	thirdPartyAddr := common.HexToAddress(thirdPartyAddress)
	if channelAddr == utils.EmptyHash || thirdPartyAddr == utils.EmptyAddress {
		err = errors.New("invalid argument")
		return
	}
	result, err := a.api.ChannelInformationFor3rdParty(channelAddr, thirdPartyAddr)
	if err != nil {
		log.Error(err.Error())
		return
	}
	r, err = marshal(result)
	return
}

/*
SwitchNetwork  switch between mesh and internet
*/
func (a *API) SwitchNetwork(isMesh bool) {
	log.Trace(fmt.Sprintf("Api SwitchNetwork isMesh=%v", isMesh))
	a.api.Raiden.Config.IsMeshNetwork = isMesh
}

/*
UpdateMeshNetworkNodes 同一个局域网内优先
*/
func (a *API) UpdateMeshNetworkNodes(nodesstr string) (err error) {
	defer func() {
		log.Trace(fmt.Sprintf("Api UpdateMeshNetworkNodes nodesstr=%s,out err=%v", nodesstr, err))
	}()
	var nodes []*network.NodeInfo
	err = json.Unmarshal([]byte(nodesstr), &nodes)
	if err != nil {
		log.Error(err.Error())
		return
	}
	err = a.api.Raiden.Protocol.UpdateMeshNetworkNodes(nodes)
	if err != nil {
		log.Error(err.Error())
		return
	}
	return nil
}

/*
EthereumStatus  query the status between raiden and ethereum
todo fix it ,r is useless
*/
func (a *API) EthereumStatus() (r string, err error) {
	c := a.api.Raiden.Chain
	if c != nil && c.Client.Status == netshare.Connected {
		return time.Now().String(), nil
	}
	return time.Now().String(), errors.New("connect failed")
}

/*
GetSentTransfers retuns list of sent transfer between `from_block` and `to_block`
*/
func (a *API) GetSentTransfers(from, to int64) (r string, err error) {
	log.Trace(fmt.Sprintf("from=%d,to=%d\n", from, to))
	trs, err := a.api.GetSentTransfers(from, to)
	if err != nil {
		log.Error(err.Error())
		return
	}
	r, err = marshal(trs)
	return
}

/*
GetReceivedTransfers retuns list of received transfer between `from_block` and `to_block`
it contains token swap
*/
func (a *API) GetReceivedTransfers(from, to int64) (r string, err error) {
	trs, err := a.api.GetReceivedTransfers(from, to)
	if err != nil {
		log.Error(err.Error())
		return
	}
	r, err = marshal(trs)
	return
}

// Subscription represents an event subscription where events are
// delivered on a data channel.
type Subscription struct {
	quitChan chan struct{}
}

// Unsubscribe cancels the sending of events to the data channel
// and closes the error channel.
func (s *Subscription) Unsubscribe() {
	close(s.quitChan)
}

// NotifyHandler is a client-side subscription callback to invoke on events and
// subscription failure.
type NotifyHandler interface {
	//some unexpected error
	OnError(errCode int, failure string)
	//OnStatusChange server connection status change
	OnStatusChange(s string)
	//OnReceivedTransfer  receive a transfer
	OnReceivedTransfer(tr string)
	//OnSentTransfer a transfer sent success
	OnSentTransfer(tr string)
}

/*
关于状态汇报,为了脱耦,单独放到一个包中,使用 channel 通信,
为了防止写阻塞,可以通过 select 写入.
向 panic一样,每次重新初始化
尽量避免 启动 go routine
如果要新创建Raiden 实例,必须调用 sub.Unsubscribe, 否则肯定会发生内存泄漏
*/

// Subscribe notifications about the current blockchain head
// on the given channel.
func (a *API) Subscribe(handler NotifyHandler) (sub *Subscription, err error) {
	sub = &Subscription{
		quitChan: make(chan struct{}),
	}
	cs := v1.ConnectionStatus{
		XMPPStatus: netshare.Disconnected,
		EthStatus:  netshare.Disconnected,
	}
	mt, ok := a.api.Raiden.Transport.(*network.MixTransporter)
	if !ok {
		err = fmt.Errorf("not MixTransporter %s", utils.StringInterface(a.api.Raiden.Transport, 3))
		return
	}
	xn, err := mt.GetNotify()
	if err != nil {
		log.Error(fmt.Sprintf("xmpp transport err %s", err))
		xn = make(chan netshare.Status)
	}
	go func() {
		rpanic.RegisterErrorNotifier("API SubscribeNeighbour")
		for {
			var err error
			var d []byte
			select {
			case err = <-rpanic.GetNotify():
				handler.OnError(32, err.Error())
			case s := <-a.api.Raiden.EthConnectionStatus:
				cs.EthStatus = s
				cs.LastBlockTime = a.api.Raiden.GetDb().GetLastBlockNumberTime().Format(v1.BlockTimeFormat)
				d, err = json.Marshal(cs)
				handler.OnStatusChange(string(d))
			case s := <-xn:
				cs.XMPPStatus = s
				cs.LastBlockTime = a.api.Raiden.GetDb().GetLastBlockNumberTime().Format(v1.BlockTimeFormat)
				d, err = json.Marshal(cs)
				handler.OnStatusChange(string(d))
			case t := <-a.api.Raiden.GetDb().SentTransferChan:
				d, err = json.Marshal(t)
				handler.OnSentTransfer(string(d))
			case t := <-a.api.Raiden.GetDb().ReceivedTransferChan:
				d, err = json.Marshal(t)
				handler.OnReceivedTransfer(string(d))
			case <-sub.quitChan:
				return
			}
			if err != nil {
				log.Error(fmt.Sprintf("err =%s", err))
			}
		}

	}()
	return
}
