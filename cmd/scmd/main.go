// Package main is the entry point for scmd
package main

import (
	"fmt"
	"os"

	"github.com/scmd/scmd/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
