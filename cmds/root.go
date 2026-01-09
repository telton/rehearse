package cmds

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/telton/rehearse/internal/logger"
	"github.com/telton/rehearse/internal/version"
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
		&cli.BoolFlag{
			Name:    "version",
			Aliases: []string{"V"},
			Usage:   "Print version information",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.Bool("version") {
			printVersion()
			return nil
		}
		return cli.ShowAppHelp(cmd)
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
		versionCmd,
	},
}

func Execute(ctx context.Context, args []string) error {
	return rootCmd.Run(ctx, args)
}

func printVersion() {
	fmt.Printf("rehearse version %s\n", version.Get())
}
