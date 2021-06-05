package interactor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fzerorubigd/bitacoin/block"
	"github.com/fzerorubigd/bitacoin/repository"
	"io/ioutil"
	"log"
	"net/http"
)

func Shout(block *block.Block) error {
	byteBlock, err := json.Marshal(block)
	if err != nil {
		return err
	}

	acceptCount := 0
	rejectCount := 0

	body := bytes.NewBuffer(byteBlock)
	for nodeAddr := range Explorer.nodes {
		request, err := http.NewRequest("POST", nodeAddr+repository.BlockUrl, body)
		if err != nil {
			log.Printf("could not create request in Shout err: %s\n", nodeAddr, err.Error())
			delete(Explorer.nodes, nodeAddr)
			continue
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			log.Printf("could not send request to %s err: %s\n", nodeAddr, err.Error())
			delete(Explorer.nodes, nodeAddr)
			continue
		}

		if response.StatusCode == http.StatusOK {
			acceptCount++
		} else {
			responseMap := make(map[string]string)
			respBody, _ := ioutil.ReadAll(response.Body)
			err = json.Unmarshal(respBody, &responseMap)
			if err != nil {
				log.Printf("unmarshal body error in shout: %s\n", err.Error())
				delete(Explorer.nodes, nodeAddr)
				continue
			}
			log.Printf("received error from node %s err: %s\n", nodeAddr, responseMap["error"])
			rejectCount++
		}
	}

	if rejectCount > acceptCount {
		return fmt.Errorf("new block is not acceptable for explored nodes, rejectCount:%d acceptCount:%d",
			rejectCount, acceptCount)
	}

	return nil
}
