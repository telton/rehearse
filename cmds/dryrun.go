package cmds

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/telton/rehearse/workflow"
)

var (
	dryRunCmd = &cli.Command{
		Name:    "dryrun",
		Aliases: []string{"dr"},
		Usage:   "analyze a workflow without running it",
		Description: `Dry-run analyzes a GitHub Actions workflow file and shows
what would run based on the current git state and simulated event.

It evaluates all conditions and shows which jobs and steps would
execute, helping you debug your workflows locally.`,
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name: "workflow-file",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "event",
				Aliases: []string{"e"},
				Usage:   "Event type to simulate (push, pull_request, etc.)",
				Value:   "push",
			},
			&cli.StringFlag{
				Name:    "ref",
				Aliases: []string{"r"},
				Usage:   "Git ref to use (defaults to current branch)",
			},
			&cli.StringSliceFlag{
				Name:    "secret",
				Aliases: []string{"s"},
				Usage:   "Secrets in KEY=VALUE format",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			workflowFile := c.StringArg("workflow-file")
			if workflowFile == "" {
				return errors.New("missing required argument: <workflow-file>")
			}

			return runDryrun(workflowFile, c.String("event"), c.String("ref"), c.StringSlice("secret"))
		},
	}
)

func runDryrun(workflowPath, eventName, ref string, secretArgs []string) error {
	wf, err := workflow.Parse(workflowPath)
	if err != nil {
		return fmt.Errorf("parsing workflow: %w", err)
	}

	secrets := make(map[string]string)
	for _, s := range secretArgs {
		secretParts := strings.SplitN(s, "=", 2)
		if len(secretParts) == 2 {
			secrets[secretParts[0]] = secretParts[1]
		}
	}

	ctx, err := workflow.NewContext(workflow.Options{
		EventName: eventName,
		Ref:       ref,
		Secrets:   secrets,
	})
	if err != nil {
		return fmt.Errorf("building context: %w", err)
	}

	a := workflow.NewAnalyzer(wf, ctx)
	result := a.Analyze()

	workflow.Render(result)

	return nil
}
