package handlers

import (
	"encoding/json"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/transaction"
	"io/ioutil"
	"log"
	"net/http"
)

var memPool []*transaction.Transaction

func TransactionHandler(w http.ResponseWriter, r *http.Request) {
	byteBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("bad new block request err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "Bad Request",
		})
		return
	}

	tnxRequest := transaction.TransactionRequest{}
	err = json.Unmarshal(byteBody, &tnxRequest)
	if err != nil {
		log.Printf("tnxRequest unmarshal err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Bad Request",
		})
		return
	}

	txn, err := blockchain.LoadedBlockChain.NewTransaction(&tnxRequest)
	if err != nil {
		log.Printf("new tnxRequest err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	log.Printf("new tnxRequest in memPool: %+v\n", tnxRequest)
	helper.WriteResponse(w, http.StatusOK, map[string]interface{}{
		"message":    "tnxRequest appended to memPool successfully",
		"tnxRequest": tnxRequest,
	})

	memPool = append(memPool, txn)
	if len(memPool) >= blockchain.LoadedBlockChain.TransactionCount {
		go func() {
			newBlock, err := blockchain.LoadedBlockChain.MineNewBlock(txn)
			if err != nil {
				log.Printf("MineNewBlock err: %s\n", err.Error())
			} else {
				log.Printf("mined new block successfully.\nlast hash: %x\nprevious hash: %x\n",
					newBlock.Hash, newBlock.PrevHash)
			}
		}()
		memPool = nil
	}
}
