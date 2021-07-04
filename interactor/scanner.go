package interactor

import (
	"fmt"
	"github.com/fzerorubigd/bitacoin/config"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/repository"
	"log"
	"net/http"
	"strings"
	"sync"
)

func Init() {
	Interactor = &interactor{
		nodes: make(map[string]struct{}),
		mutex: sync.Mutex{},
	}
	if len(config.Config.InitialNodes) < 1 {
		log.Printf("This is the first node of the decentralized network!")
	} else {
		Scan(config.Config.InitialNodes)
	}
}

type interactor struct {
	nodes map[string]struct{}
	mutex sync.Mutex
}

func (e *interactor) AddNewNode(nodeAddr string) {
	_, ok := e.nodes[nodeAddr]
	if !ok && "http://"+config.Config.Host != nodeAddr {
		log.Printf("new node has been discovered: %s\n", nodeAddr)
		e.mutex.Lock()
		defer e.mutex.Unlock()
		e.nodes[nodeAddr] = struct{}{}
	}

}

func (e *interactor) Nodes() map[string]struct{} {
	return e.nodes
}

var Interactor = &interactor{}

func Scan(initialNodes []string) {
	for i := range initialNodes {
		if initialNodes[i] == "" {
			continue
		}

		scanResp := make(map[string][]string)
		err := helper.SendReqAndUnmarshalResp(
			http.MethodGet,
			fmt.Sprintf(
				"%s%s?port=%s",
				initialNodes[i],
				repository.ExploreUrl,
				strings.Split(config.Config.Host, ":")[1],
			),
			nil,
			http.StatusOK,
			&scanResp,
		)

		if err != nil {
			log.Printf("an error happend scanning: %s\n", err.Error())
			continue
		}

		Interactor.AddNewNode(initialNodes[i])
		for i := range scanResp["nodes"] {
			Interactor.AddNewNode(scanResp["nodes"][i])
		}
	}

	if len(Interactor.nodes) < 1 {
		log.Fatalf("could not connect to any other nodes, check your config file and your network connection")
	}
}
