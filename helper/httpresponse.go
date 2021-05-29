package helper

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteResponse(w http.ResponseWriter, code int, customResponse interface{}) {
	responseByte, err := json.Marshal(customResponse)
	if err != nil {
		log.Printf("marshal bad besponse err: %s\n", err.Error())
	}
	w.WriteHeader(code)
	_, err = w.Write(responseByte)
	if err != nil {
		log.Printf("write response err: %s\n", err.Error())
	}
}
