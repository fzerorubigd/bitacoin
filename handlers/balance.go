package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/helper"
	"io/ioutil"
	"log"
	"net/http"
)

func BalanceHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("read response body err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusInternalServerError, map[string]string{
			"err": "Internal error",
		})
		return
	}

	resp := make(map[string][]byte)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Printf("unmarshal response body in balance handler err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusInternalServerError, map[string]string{
			"err": "Internal error",
		})
		return
	}

	pubKey := resp["pubKey"]
	if len(pubKey) == 0 {
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"err": "there is no pubKey in request",
		})
		return
	}

	_, balance, err := blockchain.LoadedBlockChain.UnspentTxn(pubKey)
	if err != nil {
		log.Printf("get unspent transactions err: %s\n", err.Error())
		helper.WriteResponse(w, http.StatusInternalServerError, map[string]string{
			"err": fmt.Sprintf("get unspent transactions err: %s\n", err.Error()),
		})
		return
	}

	helper.WriteResponse(w, http.StatusOK, map[string]int{
		"balance": balance,
	})
}
