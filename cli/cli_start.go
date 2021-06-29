package cli

import (
	"flag"
	"fmt"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/config"
	"github.com/fzerorubigd/bitacoin/handlers"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/interactor"
	"github.com/fzerorubigd/bitacoin/repository"
	"github.com/fzerorubigd/bitacoin/storege"
	"log"
	"net/http"
)

func start(store storege.Store, args ...string) error {
	_, err := blockchain.OpenBlockChain(repository.Difficulty, repository.TransactionCount, store)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}

	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var configFilePath string
	fs.StringVar(&configFilePath, "config", "config.json", "config file path")
	err = fs.Parse(args[1:])
	if err != nil {
		return err
	}

	err = config.ReadConfigFile(configFilePath)
	if err != nil {
		log.Fatalf("read config file err: %s\n", err.Error())
	}

	minerPubKey, err := helper.ReadKeyFromPemFile(config.Config.PubKeyPath)
	if err != nil {
		log.Fatalf("read minerPubKey failed err: %s\n", err.Error())
	}
	blockchain.LoadedBlockChain.MinerPubKey = minerPubKey

	interactor.Init()
	host := fmt.Sprintf("%s:%s", config.Config.IP, config.Config.Port)

	http.HandleFunc(repository.TransactionUrl, handlers.TransactionHandler)
	http.HandleFunc(repository.ExploreUrl, handlers.ExploreHandler)
	http.HandleFunc(repository.BlockUrl, handlers.BlockHandler)
	http.HandleFunc(repository.BalanceUrl, handlers.BalanceHandler)

	fileServer := http.FileServer(http.Dir(store.DataPath()))
	http.Handle(repository.DataServeUrl, http.StripPrefix(repository.DataServeUrl, fileServer))

	fmt.Printf("node started on host: %s\n", host)
	return http.ListenAndServe(host, nil)
}

func init() {
	addCommand("start", "start the decentralized block chain", start)
}
