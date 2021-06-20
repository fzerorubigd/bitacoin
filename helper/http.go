package helper

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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

func SendRequest(method, url string, body interface{}) ([]byte, int, error) {
	requestBody := &bytes.Buffer{}
	if body != nil {
		byteBody, err := json.Marshal(body)
		if err != nil {
			return nil, -1, err
		}
		requestBody.Write(byteBody)
	}

	request, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, -1, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, -1, err

	}

	byteResp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, -1, err
	}

	return byteResp, response.StatusCode, nil
}
