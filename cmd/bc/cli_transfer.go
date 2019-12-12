package main

import (
	"flag"
	"fmt"

	"github.com/fzerorubigd/bitacoin"
)

func transfer(store bitacoin.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		from, to string
		amount   int
	)
	fs.StringVar(&from, "from", "", "From who")
	fs.StringVar(&to, "to", "", "To who")
	fs.IntVar(&amount, "amount", 0, "amount")

	fs.Parse(args[1:])

	bc, err := bitacoin.OpenBlockChain(difficulty, store)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}

	txn, err := bitacoin.NewTransaction(bc, []byte(from), []byte(to), amount)
	if err != nil {
		return fmt.Errorf("create transaction failed: %w", err)
	}

	_, err = bc.Add(txn)

	return err
}

func init() {
	addCommand("transfer", "Transfer money", transfer)
}
