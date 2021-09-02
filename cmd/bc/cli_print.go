package main

import (
	"flag"
	"fmt"

	"github.com/fzerorubigd/bitacoin"
)

func print(store bitacoin.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		header bool
		count  int
	)
	fs.BoolVar(&header, "header", false, "Show header only")
	fs.IntVar(&count, "count", -1, "How many records to show")

	fs.Parse(args[1:])

	bc, err := bitacoin.OpenBlockChain(bitacoin.Difficulty, store)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}

	return bc.Print(header, count)
}

func init() {
	addCommand("print", "Print the chain", print)
}
