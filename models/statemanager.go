package models

import (
	"fmt"

	"github.com/SmartMeshFoundation/SmartRaiden/log"
	"github.com/SmartMeshFoundation/SmartRaiden/transfer"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"github.com/asdine/storm"
	"github.com/ethereum/go-ethereum/common"
)

const bucketAck = "ack"

//StartTx start a new tx of db
func (model *ModelDB) StartTx() (tx storm.Node) {
	var err error
	tx, err = model.db.Begin(true)
	if err != nil {
		panic(fmt.Sprintf("start transaction error %s", err))
	}
	return
}

//AddStateManager add new StateManager
func (model *ModelDB) AddStateManager(mgr *transfer.StateManager) error {
	err := model.db.Save(mgr)
	if err != nil {
		log.Error(fmt.Sprintf(" AddStateManager err=%s", err))
	}
	return err
}

//UpdateStateManaer update all fileds of StateManager
func (model *ModelDB) UpdateStateManaer(mgr *transfer.StateManager, tx storm.Node) error {
	//log.Trace(fmt.Sprintf("UpdateStateManaer %s\n", utils.StringInterface(mgr, 7)))
	err := tx.Save(mgr)
	if err != nil {
		log.Error(fmt.Sprintf("UpdateStateManaer err=%s", err))
	}
	return err
}

//GetAllStateManager return all StateManager in db
func (model *ModelDB) GetAllStateManager() []*transfer.StateManager {
	var mgrs []*transfer.StateManager
	//err := model.db.Find("IsFinished", false, &mgrs)
	err := model.db.All(&mgrs)
	if err != nil && err != storm.ErrNotFound {
		panic(fmt.Sprintf("GetAllUnfinishedStateManager err %s", err))
	}
	return mgrs
}

//GetAck get message related ack message
func (model *ModelDB) GetAck(echohash common.Hash) []byte {
	var data []byte
	err := model.db.Get(bucketAck, echohash[:], &data)
	if err != nil && err != storm.ErrNotFound {
		panic(fmt.Sprintf("GetAck err %s", err))
	}
	log.Trace(fmt.Sprintf("get ack %s from db,result=%d", utils.HPex(echohash), len(data)))
	return data
}

//SaveAck save a new ack to db
func (model *ModelDB) SaveAck(echohash common.Hash, ack []byte, tx storm.Node) {
	log.Trace(fmt.Sprintf("save ack %s to db", utils.HPex(echohash)))
	err := tx.Set(bucketAck, echohash[:], ack)
	if err != nil {
		log.Error(fmt.Sprintf("db err %s", err))
	}
}

//SaveAckNoTx save a ack to db
func (model *ModelDB) SaveAckNoTx(echohash common.Hash, ack []byte) {
	err := model.db.Set(bucketAck, echohash[:], ack)
	if err != nil {
		log.Error(fmt.Sprintf("save ack to db err %s", err))
	}
}
