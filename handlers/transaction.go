package handlers

import (
	"bytes"
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

	txnRequest := transaction.TransactionRequest{}
	err = json.Unmarshal(byteBody, &txnRequest)
	if err != nil {
		log.Printf("txnRequest unmarshal err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Bad Request",
		})
		return
	}

	txn, err := blockchain.LoadedBlockChain.NewTxn(&txnRequest)
	if err != nil {
		log.Printf("new txnRequest err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	err = blockchain.LoadedBlockChain.AddToMemPool(txn)
	if err != nil {
		helper.WriteResponse(w, http.StatusConflict, map[string]string{
			"message": err.Error(),
		})
		return
	}

	var prettyJSON bytes.Buffer
	_ = json.Indent(&prettyJSON, byteBody, "", "  ")
	log.Printf("new txnRequest in memPool:\n%s\n", prettyJSON.String())
	helper.WriteResponse(w, http.StatusOK, map[string]interface{}{
		"message":     "transaction added to memPool successfully",
		"transaction": txn,
	})
}
