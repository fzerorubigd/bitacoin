package handlers

import (
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/interactor"
	"net/http"
	"strings"
)

func ExploreHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string][]string)
	response["nodes"] = helper.ExtractKeysFromMap(interactor.Explorer.Nodes())

	port := r.URL.Query().Get("port")
	if port == "" {
		helper.WriteResponse(w, http.StatusBadRequest, map[string]string{
			"error": "port must be in url query",
		})
		return
	}

	helper.WriteResponse(w, http.StatusOK, response)

	ip := strings.Split(r.RemoteAddr, ":")[0]
	interactor.Explorer.AddNewNode(ip + ":" + port)
}