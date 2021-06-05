package cli

import (
	"flag"
	"fmt"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/config"
	"github.com/fzerorubigd/bitacoin/handlers"
	"github.com/fzerorubigd/bitacoin/storege"
	"log"
	"net/http"
)

func start(store storege.Store, args ...string) error {
	_, err := blockchain.OpenBlockChain(difficulty, transactionCount, store)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}

	err = config.ReadConfigFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("read config file err:%s\n", err.Error())
	}

	host := fmt.Sprintf("%s:%s", config.Config.IP, config.Config.Port)
	http.HandleFunc("/transaction", handlers.TransactionHandler)
	fmt.Printf("node started on host: %s\n", host)
	return http.ListenAndServe(host, nil)
}

func init() {
	addCommand("start", "start the decentralized block chain", start)
}
