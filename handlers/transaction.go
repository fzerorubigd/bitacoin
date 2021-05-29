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

type TransactionRequest struct {
	FromPubkey string
	ToPubKey   string
	Amount     int
}

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

	transaction := TransactionRequest{}
	err = json.Unmarshal(byteBody, &transaction)
	if err != nil {
		log.Printf("transaction unmarshal err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Bad Request",
		})
		return
	}

	txn, err := blockchain.LoadedBlockChain.NewTransaction([]byte(transaction.FromPubkey),
		[]byte(transaction.ToPubKey), transaction.Amount)
	if err != nil {
		log.Printf("new transaction err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	log.Printf("new transaction in memPool: %+v\n", transaction)
	helper.WriteResponse(w, http.StatusOK, map[string]interface{}{
		"message":     "transaction appended to memPool successfully",
		"transaction": transaction,
	})

	memPool = append(memPool, txn)
	if len(memPool) >= 5 {
		newBlock, err := blockchain.LoadedBlockChain.MineNewBlock(txn)
		if err != nil {
			log.Printf("MineNewBlock err: %s\n", err.Error())
			helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		} else {
			log.Printf("mined new block successfully.\nlast hash: %x\nprevious hash: %x\n",
				newBlock.Hash, newBlock.PrevHash)
		}
		memPool = nil
	}
}
