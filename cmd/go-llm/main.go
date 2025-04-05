package main

import (
	"fmt"
	"os"

	"github.com/zerobang-dev/go-llm/cmd/go-llm/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
