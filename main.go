package main

import (
	"os"

	"github.com/sentient/statsdbeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
