package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()
	bc := NewBlockChain(4)
	bc.Add("Hello")
	bc.Add("Another")

	if err := bc.Validate(); err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Println(bc)
}
