package cli

import (
	"flag"
	"fmt"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/storege"
)

const difficulty = 2

func balance(store storege.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		owner string
	)
	fs.StringVar(&owner, "owner", "", "Who?")

	fs.Parse(args[1:])

	bc, err := blockchain.OpenBlockChain(difficulty, store)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}

	_, _, acc, err := bc.UnspentTxn([]byte(owner))
	if err != nil {
		return fmt.Errorf("get balance failed: %w", err)
	}

	fmt.Printf("The balance for %s is %d", owner, acc)

	return nil
}

func init() {
	addCommand("balance", "Print balance for someone", balance)
}
