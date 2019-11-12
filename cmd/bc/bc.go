package main

import (
	"flag"
	"log"
	"os"

	"github.com/fzerorubigd/bitacoin"
)

const (
	difficulty = 2
)

func main() {
	var store string
	flag.StringVar(&store, "store", os.Getenv("BC_STORE"), "The store to use")
	flag.Usage = usage
	flag.Parse()

	s := bitacoin.NewFolderStore(store)

	if err := dispatch(s, flag.Args()...); err != nil {
		log.Fatal(err.Error())
	}
}
