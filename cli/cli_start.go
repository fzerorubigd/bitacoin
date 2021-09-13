package cli

import (
	"flag"
	"fmt"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/config"
	"github.com/fzerorubigd/bitacoin/handlers"
	"github.com/fzerorubigd/bitacoin/interactor"
	"github.com/fzerorubigd/bitacoin/repository"
	"github.com/fzerorubigd/bitacoin/storege"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func start(store storege.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var configFilePath string
	fs.StringVar(&configFilePath, "config", "config.json", "config file path")
	err := fs.Parse(args[1:])
	if err != nil {
		return err
	}

	err = config.ReadConfigFile(configFilePath)
	if err != nil {
		log.Fatalf("read config file err: %s\n", err.Error())
	}

	err = blockchain.OpenBlockChain(repository.Difficulty, repository.TransactionCount, store)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}

	http.HandleFunc(repository.TransactionUrl, handlers.TransactionHandler)
	http.HandleFunc(repository.ExploreUrl, handlers.ExploreHandler)
	http.HandleFunc(repository.BlockUrl, handlers.BlockHandler)
	http.HandleFunc(repository.BalanceUrl, handlers.BalanceHandler)

	fileServer := http.FileServer(http.Dir(store.DataPath()))
	http.Handle(repository.DataServeUrl, http.StripPrefix(repository.DataServeUrl, fileServer))

	go func() {
		fmt.Printf("node started on host: %s\n", config.Config.Host)
		err := http.ListenAndServe(config.Config.Host, nil)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}()

	<-time.After(time.Second)

	err = interactor.Init()
	if err != nil {
		log.Printf("intract err: %s\n", err.Error())
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	return nil
}

func init() {
	addCommand("start", "start the decentralized blockchain", start)
}
