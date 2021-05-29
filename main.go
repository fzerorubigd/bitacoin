package main

import (
	"flag"
	"github.com/fzerorubigd/bitacoin/cli"
	"github.com/fzerorubigd/bitacoin/storege"
	"log"
	"os"
)

func main() {
	var storePath string
	flag.StringVar(&storePath, "store", os.Getenv("BC_STORE"), "The store to use")
	flag.Usage = cli.Usage
	flag.Parse()

	s := storege.NewFolderStore(storePath)

	if err := cli.Dispatch(s, flag.Args()...); err != nil {
		log.Fatal(err.Error())
	}
}
