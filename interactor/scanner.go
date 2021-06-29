package interactor

import (
	"fmt"
	"github.com/fzerorubigd/bitacoin/config"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/repository"
	"log"
	"net/http"
	"sync"
)

func Init() {
	Explorer = &explorer{
		nodes: make(map[string]struct{}),
		mutex: sync.Mutex{},
	}
	if len(config.Config.InitialNodes) < 1 {
		log.Printf("This is the first node of the decentralized network!")
	} else {
		Scan(config.Config.InitialNodes)
	}
}

type explorer struct {
	nodes map[string]struct{}
	mutex sync.Mutex
}

func (e *explorer) AddNewNode(nodeAddr string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.nodes[nodeAddr] = struct{}{}
}

func (e *explorer) Nodes() map[string]struct{} {
	return e.nodes
}

var Explorer = &explorer{}

func Scan(initialNodes []string) {
	for i := range initialNodes {
		if initialNodes[i] == "" {
			continue
		}

		scanResp := make(map[string][]string)
		err := helper.SendReqAndUnmarshalResp(
			http.MethodGet,
			fmt.Sprintf("%s%s?port=%s", initialNodes[i], repository.ExploreUrl, config.Config.Port),
			nil,
			http.StatusOK,
			&scanResp,
		)

		if err != nil {
			log.Printf("an error happend scanning: %s\n", err.Error())
			continue
		}

		Explorer.AddNewNode(initialNodes[i])
		for i := range scanResp["nodes"] {
			Explorer.AddNewNode(scanResp["nodes"][i])
		}
	}

	if len(Explorer.nodes) < 1 {
		log.Fatalf("could not connect to any other nodes, check your config file and your network connection")
	}
}
