package main

import (
	"flag"
	"github.com/fzerorubigd/bitacoin"
	"log"
	"os"
)

const (
	difficulty = 20
)

func main() {

	flag.Usage = usage
	flag.Parse()

	runBoltDB()
}

func runByFile() {
	var store string
	flag.StringVar(&store, "store", os.Getenv("BC_STORE"), "The store to use")
	s := bitacoin.NewFolderStore(store)

	if err := dispatch(s, flag.Args()...); err != nil {
		log.Fatal(err.Error())
	}
}

func runBoltDB() {

	d := bitacoin.NewDBStore()
	if err := dispatch(d, flag.Args()...); err != nil {
		log.Fatal(err.Error())
	}
}
