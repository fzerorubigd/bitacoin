package main

import (
	"flag"
	"github.com/fzerorubigd/bitacoin"
	"log"
)

const (
	difficulty = 20
)

func main() {
	//var store string
	//flag.StringVar(&store, "store", os.Getenv("BC_STORE"), "The store to use")
	flag.Usage = usage
	flag.Parse()

	//s := bitacoin.NewFolderStore(store)
	//
	//if err := dispatch(s, flag.Args()...); err != nil {
	//	log.Fatal(err.Error())
	//}

	d := bitacoin.NewDBStore()
	if err := dispatch(d, flag.Args()...); err != nil {
		log.Fatal(err.Error())
	}
}
