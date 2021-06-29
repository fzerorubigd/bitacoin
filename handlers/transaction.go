package handlers

import (
	"context"
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

	log.Println("new tnxRequest in memPool")
	helper.WriteResponse(w, http.StatusOK, map[string]interface{}{
		"message":    "tnxRequest appended to memPool successfully",
		"tnxRequest": tnxRequest,
	})

	memPool = append(memPool, txn)
	if len(memPool) >= blockchain.LoadedBlockChain.TransactionCount && blockchain.LoadedBlockChain.CancelMining == nil {
		ctx, cancel := context.WithCancel(context.Background())
		blockchain.LoadedBlockChain.CancelMining = cancel

		transactions := memPool[:blockchain.LoadedBlockChain.TransactionCount]
		memPool = memPool[blockchain.LoadedBlockChain.TransactionCount:]
		go func() {
			newBlock, err := blockchain.LoadedBlockChain.StartMining(ctx, transactions...)
			if err != nil {
				log.Printf("StartMining err: %s\n", err.Error())
			} else {
				log.Printf("new block added to the blockchain successfully.\n%s\n", newBlock.String())
			}
			blockchain.LoadedBlockChain.CancelMining = nil
		}()
		memPool = nil
	}
}
