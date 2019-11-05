package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fzerorubigd/bitacoin"
)

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()
	bc, err := bitacoin.NewBlockChain(4, bitacoin.NewFolderStore("/home/f0rud/bitac"))
	if err != nil {
		log.Fatal(err.Error())
	}
	bc.Add("Hello")
	bc.Add("Another")

	if err := bc.Validate(); err != nil {
		log.Fatalf(err.Error())
	}

	bc.Print()
}
