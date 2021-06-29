package cli

import (
	"flag"
	"fmt"
	"github.com/fzerorubigd/bitacoin/blockchain"
	"github.com/fzerorubigd/bitacoin/repository"
	"github.com/fzerorubigd/bitacoin/storege"
)

func validate(store storege.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)

	fs.Parse(args[1:])

	bc, err := blockchain.OpenBlockChain(repository.Difficulty, repository.TransactionCount, store)
	if err != nil {
		return fmt.Errorf("open failed: %w", err)
	}

	return bc.Validate()
}

func init() {
	addCommand("validate", "Validate the blockchain", validate)
}
