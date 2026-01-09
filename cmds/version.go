package cmds

import (
	"context"

	"github.com/urfave/cli/v3"
)

var versionCmd = &cli.Command{
	Name:    "version",
	Usage:   "Show version information",
	Aliases: []string{"v"},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		printVersion()
		return nil
	},
}