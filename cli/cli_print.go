package cli

import (
	"flag"
	"fmt"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/repository"
	"github.com/fzerorubigd/bitacoin/storege"
)

func print(store storege.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		header bool
		count  int
	)
	fs.BoolVar(&header, "header", false, "Show header only")
	fs.IntVar(&count, "count", -1, "How many records to show")

	fs.Parse(args[1:])

	bc, err := blockchain.OpenBlockChain(repository.Difficulty, repository.TransactionCount, store)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}

	return bc.Print(header, count)
}

func init() {
	addCommand("print", "Print the chain", print)
}
