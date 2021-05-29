package handlers

import (
	"encoding/json"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/helper"
	"io/ioutil"
	"log"
	"net/http"
)

type TransactionRequest struct {
	FromPubkey string
	ToPubKey   string
	Amount     int
}

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

	_, err = blockchain.LoadedBlockChain.Add(txn)
	if err != nil {
		log.Printf("add transaction err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	log.Printf("successfull transaction: %+v\n", transaction)
	helper.WriteResponse(w, http.StatusOK, map[string]interface{}{
		"message":     "transferred successfully",
		"transaction": transaction,
	})
}
