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
	bc := bitacoin.NewBlockChain(4)
	bc.Add("Hello")
	bc.Add("Another")

	if err := bc.Validate(); err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Println(bc)
}
