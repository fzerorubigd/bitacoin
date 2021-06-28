package cli

import (
	"flag"
	"fmt"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/storege"
	"log"
)

func initialize(store storege.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	walletDir := fs.String("wallet", "./wallet", "initial node wallet directory path")
	_ = fs.Parse(args[1:])

	_, pubKey, err := helper.GenerateWallet(*walletDir)
	if err != nil {
		return fmt.Errorf("generate wallet err: %s", err.Error())
	}

	_, err = blockchain.NewBlockChain(pubKey, difficulty, transactionCount, store)
	if err != nil {
		return fmt.Errorf("create new blockchain failed: %w", err)
	}

	log.Println("blockchain initialized successfully")

	return nil
}

func init() {
	addCommand("init", "Create an empty blockchain", initialize)
}
