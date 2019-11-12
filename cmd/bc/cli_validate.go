package main

import (
	"flag"
	"fmt"

	"github.com/fzerorubigd/bitacoin"
)

func validate(store bitacoin.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)

	fs.Parse(args[1:])

	bc, err := bitacoin.OpenBlockChain(difficulty, store)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}

	return bc.Validate()
}

func init() {
	addCommand("validate", "Validate the blockchain", validate)
}
