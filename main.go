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
	flag.StringVar(&storePath, "store", "./data", "The store to use")

	err := os.MkdirAll(storePath, 0666)
	if err != nil {
		log.Fatalf("mkdirAll failed, err: %s\n", err.Error())
	}

	flag.Usage = cli.Usage
	flag.Parse()

	s := storege.NewFolderStore(storePath)

	if err := cli.Dispatch(s, flag.Args()...); err != nil {
		log.Fatal(err.Error())
	}
}
