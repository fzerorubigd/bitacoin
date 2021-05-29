package cli

import (
	"flag"
	"fmt"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/storege"
)

func initialize(store storege.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		genesis string
	)
	fs.StringVar(&genesis, "owner", "bita", "Genesis block owner")

	fs.Parse(args[1:])

	_, err := blockchain.NewBlockChain([]byte(genesis), difficulty, store)
	if err != nil {
		return fmt.Errorf("create failed: %w", err)
	}
	return nil
}

func init() {
	addCommand("init", "Create an empty blockchain", initialize)
}
