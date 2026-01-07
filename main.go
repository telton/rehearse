package main

import (
	"context"
	"fmt"
	"os"

	"github.com/telton/rehearse/cmds"
)

func main() {
	ctx := context.Background()

	if err := cmds.Execute(ctx, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
