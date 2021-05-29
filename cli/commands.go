package cli

import (
	"flag"
	"fmt"
	"github.com/fzerorubigd/bitacoin/storege"
	"os"
)

var (
	allCommand []command
)

type command struct {
	Name        string
	Description string

	Run func(storege.Store, ...string) error
}

func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()

	fmt.Fprintf(flag.CommandLine.Output(), "Sub commands:\n")

	for i := range allCommand {
		fmt.Fprintf(flag.CommandLine.Output(), "  %s: %s\n", allCommand[i].Name, allCommand[i].Description)
	}

}

func Dispatch(store storege.Store, args ...string) error {
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

func addCommand(name, description string, run func(storege.Store, ...string) error) {
	allCommand = append(allCommand, command{
		Name:        name,
		Description: description,
		Run:         run,
	})
}
