package cli

import (
	"flag"
	"fmt"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/repository"
	"github.com/fzerorubigd/bitacoin/storege"
	"log"
)

func initialize(store storege.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	pubKeyDir := fs.String("pub", "", "wallet public key of the first transaction and genesis block")
	_ = fs.Parse(args[1:])

	pubKey, err := helper.ReadKeyFromPemFile(*pubKeyDir)
	if err != nil {
		log.Fatalf("read minerPubKey failed err: %s\n", err.Error())
	}

	_, err = blockchain.NewBlockChain(pubKey, repository.Difficulty, repository.TransactionCount, store)
	if err != nil {
		return fmt.Errorf("create new blockchain failed: %w", err)
	}

	log.Println("blockchain initialized successfully")
	return nil
}

func init() {
	addCommand("init", "Create an empty blockchain", initialize)
}
