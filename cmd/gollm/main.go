package main

import (
	"fmt"
	"os"

	"github.com/zerobang-dev/gollm/cmd/gollm/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
