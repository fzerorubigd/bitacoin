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

func init() {
	Explore(config.Config.InitialNodes)
	if len(Explorer.nodes) < 1 {
		log.Fatalf("no nodes found!!! check initialNodes.json and network connection")
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

func (e *explorer) Nodes() []string {
	nodes := make([]string, len(e.nodes))
	for node := range e.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

var Explorer = &explorer{}

func Explore(initialNodes []string) {
	for i := range initialNodes {
		response, err := http.Get(fmt.Sprintf("%s/%s?port=%d", initialNodes[i], repository.ExploreUrl,
			config.Config.Port))
		if err != nil {
			log.Printf("could not send Explore request to node %s err: %s\n", initialNodes[i], err.Error())
			continue
		}

		if response.StatusCode == http.StatusOK {
			exploreResponse := make(map[string][]string)
			respBody, _ := ioutil.ReadAll(response.Body)
			err = json.Unmarshal(respBody, &exploreResponse)
			if err != nil {
				log.Printf("unmarshal body error in Explore: %s\n", err.Error())
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
}
