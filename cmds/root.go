package cmds

import (
	"context"

	"github.com/urfave/cli/v3"
)

var rootCmd = &cli.Command{
	Name:  "rehearse",
	Usage: "practice before the real thing",
	Commands: []*cli.Command{
		dryRunCmd,
		listCmd,
		runCmd,
	},
}

func Execute(ctx context.Context, args []string) error {
	return rootCmd.Run(ctx, args)
}
