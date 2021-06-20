package interactor

import (
	"encoding/json"
	"fmt"
	"github.com/fzerorubigd/bitacoin/config"
	"github.com/fzerorubigd/bitacoin/repository"
	"io/ioutil"
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
		log.Printf("This is the first node of the cryptocurrenct network!")
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

		request, err := http.NewRequest("GET", fmt.Sprintf("%s%s", initialNodes[i], repository.ExploreUrl),
			nil)
		if err != nil {
			log.Printf("could not create request err: %s\n", err.Error())
			continue
		}

		query := request.URL.Query()
		query.Add("port", config.Config.Port)
		request.URL.RawQuery = query.Encode()

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			log.Printf("could not scan node %s err: %s\n", initialNodes[i], err.Error())
			continue
		}

		if response.StatusCode == http.StatusOK {
			exploreResponse := make(map[string][]string)
			respBody, _ := ioutil.ReadAll(response.Body)
			err = json.Unmarshal(respBody, &exploreResponse)
			if err != nil {
				log.Printf("unmarshal body error in Scan: %s\n", err.Error())
				continue
			}
			Explorer.AddNewNode(initialNodes[i])
			for i := range exploreResponse["nodes"] {
				Explorer.AddNewNode(exploreResponse["nodes"][i])
			}
		} else {
			responseMap := make(map[string]string)
			respBody, _ := ioutil.ReadAll(response.Body)
			err = json.Unmarshal(respBody, &responseMap)
			if err != nil {
				log.Printf("unmarshal body error in shout: %s\n", err.Error())
				continue
			}
			log.Printf("received error from node %s err: %s\n", initialNodes[i], responseMap["error"])
		}
	}

	if len(Explorer.nodes) < 1 {
		log.Fatalf("could not connect to any other nodes, check your config file and your network connection")
	}
}
