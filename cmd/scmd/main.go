// Package main is the entry point for scmd
package main

import (
	"os"

	"github.com/scmd/scmd/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
