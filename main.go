package main

import (
	"os"

	"github.com/sentient/statsdbeat/cmd"

	_ "github.com/sentient/statsdbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
