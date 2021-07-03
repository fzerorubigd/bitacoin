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

	log.Println("new tnxRequest in memPool")
	helper.WriteResponse(w, http.StatusOK, map[string]interface{}{
		"message":     "transaction added to memPool successfully",
		"transaction": txn,
	})

	blockchain.LoadedBlockChain.AddToMemPool(txn)
}
