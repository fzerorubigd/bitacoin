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
		log.Printf("block unmarshal err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Bad Request",
		})
		return
	}

	err = newBlock.Validate(blockchain.LoadedBlockChain.Mask)
	if err != nil {
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Block",
		})
		return
	}

	err = blockchain.LoadedBlockChain.AppendBlock(&newBlock)
	if err != nil {
		log.Printf("append new block err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	log.Printf("new block Accepted: %+v\n", newBlock)
	helper.WriteResponse(w, http.StatusOK, map[string]interface{}{
		"message": "new block appended successfully",
		"block":   newBlock,
	})
}
