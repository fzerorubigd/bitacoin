package interactor

import (
	"fmt"
	"github.com/fzerorubigd/bitacoin/block"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/repository"
	"log"
	"net/http"
)

func Shout(block *block.Block) error {
	log.Println("shouting started")

	acceptCount := 0
	rejectCount := 0

	for nodeAddr := range Interactor.nodes {
		shoutResp := make(map[string]string)
		err := helper.SendReqAndUnmarshalResp(
			http.MethodPost,
			nodeAddr+repository.BlockUrl,
			block,
			http.StatusOK,
			&shoutResp,
		)
		if err != nil {
			log.Printf("an error happend while shouting: %s\n", err.Error())
			rejectCount++
			continue
		}

		acceptCount++
	}

	if rejectCount > acceptCount {
		return fmt.Errorf("new block is not acceptable for explored nodes, rejectCount:%d acceptCount:%d",
			rejectCount, acceptCount)
	}

	return nil
}
