package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fzerorubigd/bitacoin"
)

var (
	allCommand []command
)

type command struct {
	Name        string
	Description string

	Run func(bitacoin.Store, ...string) error
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()

	fmt.Fprintf(flag.CommandLine.Output(), "Sub commands:\n")

	for i := range allCommand {
		fmt.Fprintf(flag.CommandLine.Output(), "  %s: %s\n", allCommand[i].Name, allCommand[i].Description)
	}

}

func dispatch(store bitacoin.Store, args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("atleast one arg is required")
	}

	sub := args[0]
	for i := range allCommand {
		if sub == allCommand[i].Name {
			return allCommand[i].Run(store, args...)
		}
	}

	flag.Usage()
	return fmt.Errorf("invalid command")
}

func addCommand(name, description string, run func(bitacoin.Store, ...string) error) {
	allCommand = append(allCommand, command{
		Name:        name,
		Description: description,
		Run:         run,
	})
}
