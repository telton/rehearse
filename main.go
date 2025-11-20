package main

import (
	"os"

	"github.com/telton/rehearse/cmds"
)

func main() {
	if err := cmds.Execute(); err != nil {
		os.Exit(1)
	}
}
