package main

import (
	"flag"
	"fmt"

	"github.com/fzerorubigd/bitacoin"
)

func initialize(store bitacoin.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		genesis string
	)
	fs.StringVar(&genesis, "genesis", "Genesis block", "Genesis data to use")

	fs.Parse(args[1:])

	_, err := bitacoin.NewBlockChain(genesis, difficulty, store)
	if err != nil {
		return fmt.Errorf("create failed: %w", err)
	}
	return nil
}

func init() {
	addCommand("init", "Create an empty blockchain", initialize)
}
