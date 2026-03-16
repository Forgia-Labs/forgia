package main

import (
	"os"

	"github.com/Deepzima/forgia/cmd/forgia/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
