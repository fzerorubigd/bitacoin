package handlers

import (
	"encoding/json"
	"github.com/fzerorubigd/bitacoin/block"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/helper"
	"io/ioutil"
	"log"
	"net/http"
)

func BlockHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("received new block from an other node")

	byteBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("bad new block request err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "Bad Request",
		})
		return
	}

	newBlock := block.Block{}
	err = json.Unmarshal(byteBody, &newBlock)
	if err != nil {
		log.Printf("block unmarshal, err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Bad Request",
		})
		return
	}

	err = newBlock.Validate(blockchain.LoadedBlockChain.Mask)
	if err != nil {
		log.Printf("recieved invalid block, err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Block",
		})
		return
	}
	log.Println("block nonce is valid")

	err = blockchain.LoadedBlockChain.ValidateIncomingTransactions(newBlock.Transactions)
	if err != nil {
		log.Printf("recieved valid block but with invalid transactions, err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Transactions",
		})
		return
	}
	log.Println("block transactions are valid")

	if blockchain.LoadedBlockChain.CancelMining != nil {
		blockchain.LoadedBlockChain.CancelMining()
		log.Println("mining canceled")
	}

	err = blockchain.LoadedBlockChain.AppendBlock(&newBlock)
	if err != nil {
		log.Printf("append new block, err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	log.Printf("new block Accepted:\n%s\n", newBlock.String())
	helper.WriteResponse(w, http.StatusOK, map[string]interface{}{
		"message": "new block appended successfully",
	})
}
