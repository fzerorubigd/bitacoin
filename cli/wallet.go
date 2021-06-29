package cli

import (
	"flag"
	"fmt"
	"github.com/fzerorubigd/bitacoin/helper"
	"github.com/fzerorubigd/bitacoin/storege"
	"log"
)

func wallet(store storege.Store, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	dirPath := fs.String("dir", "new_wallet", "wallet key pairs directory path")
	err := fs.Parse(args[1:])
	if err != nil {
		return err
	}

	if dirPath == nil {
		return fmt.Errorf("dir parameter is required:")
	}

	_, _, err = helper.GenerateWallet(*dirPath)
	if err != nil {
		return fmt.Errorf("generate new wallet failed, err: %s\n", err.Error())
	}

	log.Println("wallet generated successfully")
	return nil
}

func init() {
	addCommand("wallet", "generate new wallet", wallet)
}
