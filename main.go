package main

import (
	"os"

	"github.com/nachoal/gwt/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
