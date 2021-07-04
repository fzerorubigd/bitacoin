package downloader

import (
	"fmt"
	"github.com/fzerorubigd/bitacoin/block"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/interactor"
	"github.com/fzerorubigd/bitacoin/repository"
	"github.com/fzerorubigd/bitacoin/storege"
	"log"
	"net/http"
)

func DownloadBlockChainData(store storege.Store) {
	nodes := interactor.Interactor.Nodes()
	lastHash := getLastHashFromOtherNodes(nodes)
	if len(lastHash) == 0 {
		log.Fatalf("could not get last hash from other nodes, check your network connection")
	}

	err := store.WriteJSON("lastHash.json", map[string]interface{}{"lastHash": lastHash})
	if err != nil {
		log.Fatalf("write last hash in file err: %s\n", err.Error())
	}

	blockFileName := fmt.Sprintf("%x.json", lastHash)

Finished:
	for {
		if len(nodes) <= 0 {
			log.Printf("there is not any node for download the blockchain")
			break Finished
		}

		for nodeUrl := range nodes {
			for i := 0; i < 4; i++ {
				newBlock, err := downloadBlock(fmt.Sprintf("%s%s%s", nodeUrl, repository.DataServeUrl, blockFileName))
				if err != nil {
					log.Printf("got an error while downloading the blockchain, err: %s\n", err.Error())
					delete(nodes, nodeUrl)
					break
				}

				err = store.WriteJSON(blockFileName, newBlock)
				if err != nil {
					log.Printf("got an error while writing new block, err: %s\n", err.Error())
					delete(nodes, nodeUrl)
					break
				}

				if len(newBlock.PrevHash) > 0 {
					blockFileName = fmt.Sprintf("%x", newBlock.PrevHash) + ".json"
				} else {
					log.Println("blockchain downloaded successfully")
					break Finished
				}
			}
		}
	}
}

func downloadBlock(url string) (*block.Block, error) {
	newBlock := &block.Block{}
	err := helper.SendReqAndUnmarshalResp(
		http.MethodGet,
		url,
		nil,
		http.StatusOK,
		newBlock,
	)
	if err != nil {
		return nil, fmt.Errorf("donwload block err: %w", err.Error())
	}

	return newBlock, nil
}

func getLastHashFromOtherNodes(nodes map[string]struct{}) []byte {
	for nodeUrl := range nodes {
		lastHash, err := getLastHash(fmt.Sprintf("%s%s%s", nodeUrl, repository.DataServeUrl, "lastHash.json"))
		if err != nil {
			log.Printf("got an error while getting last hash: %s\n", err.Error())
		} else {
			return lastHash
		}
	}

	return nil
}

func getLastHash(url string) ([]byte, error) {
	respMap := make(map[string][]byte)
	err := helper.SendReqAndUnmarshalResp(
		http.MethodGet,
		url,
		nil,
		http.StatusOK,
		&respMap,
	)
	if err != nil {
		return nil, fmt.Errorf("get lastHash err: %w", err.Error())
	}

	if len(respMap["lastHash"]) == 0 {
		return nil, fmt.Errorf(`recived bad response from node %s, err: there is no lastBlock in response body,
recieved response: %+v`, url, respMap)
	}

	return respMap["lastHash"], nil
}
