package v1

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/SmartMeshFoundation/SmartRaiden/log"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ethereum/go-ethereum/common"
)

/*
TokenSwap is the api of /api/1/tokenswap/:id
:id must be a unique identifier.
*/
func TokenSwap(w rest.ResponseWriter, r *rest.Request) {
	/*
	   {
	       "role": "maker",
	       "sending_amount": 42,
	       "sending_token": "0xea674fdde714fd979de3edf0f56aa9716b898ec8",
	       "receiving_amount": 76,
	       "receiving_token": "0x2a65aca4d5fc5b5c859090a6c34d164135398226"
	   }
	*/
	type Req struct {
		Role            string   `json:"role"`
		SendingAmount   *big.Int `json:"sending_amount"`
		SendingToken    string   `json:"sending_token"`
		ReceivingAmount *big.Int `json:"receiving_amount"`
		ReceivingToken  string   `json:"receiving_token"`
	}
	targetstr := r.PathParam("target")
	idstr := r.PathParam("id")
	var target common.Address
	var id int
	if len(targetstr) != len(target.String()) {
		rest.Error(w, "target address error", http.StatusBadRequest)
		return
	}
	target = common.HexToAddress(targetstr)
	id, err := strconv.Atoi(idstr)
	if id <= 0 || err != nil {
		rest.Error(w, "must provide a valid id ", http.StatusBadRequest)
		return
	}
	req := &Req{}
	err = r.DecodeJsonPayload(req)
	if err != nil {
		log.Error(err.Error())
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Role == "maker" {
		err = RaidenAPI.TokenSwapAndWait(strconv.Itoa(id), common.HexToAddress(req.SendingToken), common.HexToAddress(req.ReceivingToken),
			RaidenAPI.Raiden.NodeAddress, target, req.SendingAmount, req.ReceivingAmount)
	} else if req.Role == "taker" {
		err = RaidenAPI.ExpectTokenSwap(strconv.Itoa(id), common.HexToAddress(req.ReceivingToken), common.HexToAddress(req.SendingToken),
			target, RaidenAPI.Raiden.NodeAddress, req.ReceivingAmount, req.SendingAmount)
	} else {
		err = fmt.Errorf("Provided invalid token swap role %s", req.Role)
	}
	if err != nil {
		log.Error(err.Error())
		rest.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		w.(http.ResponseWriter).WriteHeader(http.StatusCreated)
		_, err = w.(http.ResponseWriter).Write(nil)
		if err != nil {
			log.Warn(fmt.Sprintf("writejson err %s", err))
		}
	}
}
