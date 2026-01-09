package cmds

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/telton/rehearse/workflow"
)

var (
	runCmd = &cli.Command{
		Name:    "run",
		Aliases: []string{"r"},
		Usage:   "execute a workflow locally using Docker",
		Description: `Run executes a GitHub Actions workflow file locally using Docker containers.

It parses the workflow, evaluates conditions, and executes jobs and steps
in Docker containers, providing a local environment for testing workflows.

This command supports:
- Shell commands (run:)
- GitHub Actions (uses:) - local, repository, and docker actions
- Environment variables and secrets
- Job dependencies and conditions
- Most GitHub Actions syntax

Requirements:
- Docker must be installed and running
- Git must be available for cloning actions`,
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
			&cli.StringFlag{
				Name:  "working-dir",
				Usage: "Working directory for workflow execution (defaults to current directory)",
				Value: ".",
			},
			&cli.BoolFlag{
				Name:  "pull",
				Usage: "Always pull Docker images before running",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "cleanup",
				Usage: "Clean up containers and volumes after execution",
				Value: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			workflowFile := c.StringArg("workflow-file")
			if workflowFile == "" {
				return errors.New("missing required argument: <workflow-file>")
			}

			return runWorkflow(ctx, runConfig{
				WorkflowFile: workflowFile,
				EventName:    c.String("event"),
				Ref:          c.String("ref"),
				SecretArgs:   c.StringSlice("secret"),
				WorkingDir:   c.String("working-dir"),
				Pull:         c.Bool("pull"),
				Cleanup:      c.Bool("cleanup"),
			})
		},
	}
)

// runConfig holds configuration for workflow execution.
type runConfig struct {
	WorkflowFile string
	EventName    string
	Ref          string
	SecretArgs   []string
	WorkingDir   string
	Pull         bool
	Cleanup      bool
}

// runWorkflow executes a workflow with the given configuration.
func runWorkflow(ctx context.Context, config runConfig) error {
	renderer := workflow.NewRunRenderer()

	workingDir, err := filepath.Abs(config.WorkingDir)
	if err != nil {
		return fmt.Errorf("resolving working directory: %w", err)
	}

	if _, err := os.Stat(workingDir); os.IsNotExist(err) {
		return fmt.Errorf("working directory does not exist: %s", workingDir)
	}

	wf, err := workflow.Parse(config.WorkflowFile)
	if err != nil {
		return fmt.Errorf("parsing workflow: %w", err)
	}

	secrets := make(map[string]string)
	for _, s := range config.SecretArgs {
		secretParts := strings.SplitN(s, "=", 2)
		if len(secretParts) == 2 {
			secrets[secretParts[0]] = secretParts[1]
		}
	}

	triggerContext, err := workflow.NewContext(workflow.Options{
		EventName: config.EventName,
		Ref:       config.Ref,
		Secrets:   secrets,
	})
	if err != nil {
		return fmt.Errorf("building context: %w", err)
	}

	renderer.RenderDockerCheck()
	if err := validateDockerAvailable(); err != nil {
		renderer.RenderDockerError(err)
		return err
	}
	renderer.RenderDockerSuccess()

	renderer.RenderDockerInit()
	dockerClient, err := workflow.NewDockerClient()
	if err != nil {
		return fmt.Errorf("initializing Docker client: %w", err)
	}
	defer func() {
		if closer, ok := dockerClient.(interface{ Close() error }); ok {
			closer.Close()
		}
	}()

	gitClient := workflow.NewGitRepo()

	analyzer := workflow.NewAnalyzer(wf, triggerContext)

	executor := workflow.NewExecutor(analyzer, dockerClient, gitClient)
	executor.SetWorkingDirectory(workingDir)

	renderer.RenderWorkflowStart(wf.Name, workingDir, config.EventName, config.Ref)

	renderer.RenderExecutionStart()
	if err := executor.Execute(ctx, wf, triggerContext); err != nil {
		renderer.RenderWorkflowError(err)
		return fmt.Errorf("executing workflow: %w", err)
	}

	renderer.RenderWorkflowSuccess()
	return nil
}

// validateDockerAvailable checks if Docker is available and running.
func validateDockerAvailable() error {
	dockerClient, err := workflow.NewDockerClient()
	if err != nil {
		return fmt.Errorf("docker is not available: %w", err)
	}
	defer func() {
		if closer, ok := dockerClient.(interface{ Close() error }); ok {
			closer.Close()
		}
	}()

	if realClient, ok := dockerClient.(*workflow.RealDockerClient); ok {
		ctx := context.Background()
		_, err = realClient.Ping(ctx)
		if err != nil {
			return fmt.Errorf("docker daemon is not responding: %w", err)
		}
	}

	return nil
}
