package cmds

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/telton/rehearse/internal/logger"
)

var rootCmd = &cli.Command{
	Name:  "rehearse",
	Usage: "practice before the real thing",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "log-level",
			Aliases: []string{"l"},
			Usage:   "Set log level (debug, info, warn, error)",
			Value:   "info",
			Sources: cli.EnvVars("REHEARSE_LOG_LEVEL"),
		},
	},
	Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		// Setup logger with the specified level
		logLevel := cmd.String("log-level")
		level := logger.ParseLevelFromString(logLevel)

		cfg := &logger.Config{
			Level:  level,
			Format: "text",
			Output: os.Stdout,
		}

		logger.Setup(cfg)
		return ctx, nil
	},
	Commands: []*cli.Command{
		dryRunCmd,
		listCmd,
		runCmd,
	},
}

func Execute(ctx context.Context, args []string) error {
	return rootCmd.Run(ctx, args)
}
