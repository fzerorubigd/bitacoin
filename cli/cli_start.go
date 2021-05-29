package cli

import (
	"flag"
	"fmt"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/handlers"
	"github.com/fzerorubigd/bitacoin/storege"
	"net/http"
)

func start(store storege.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var port string
	var address string
	fs.StringVar(&port, "port", "9090", "http port")
	fs.StringVar(&address, "ip", "127.0.0.1", "ip address")
	fs.Parse(args[1:])

	_, err := blockchain.OpenBlockChain(difficulty, transactionCount, store)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}

	host := fmt.Sprintf("%s:%s", address, port)
	http.HandleFunc("/transaction", handlers.TransactionHandler)
	fmt.Printf("node started on host: %s\n", host)
	return http.ListenAndServe(host, nil)
}

func init() {
	addCommand("start", "start the decentralized block chain", start)
}
