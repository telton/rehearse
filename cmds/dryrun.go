package cmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/telton/rehearse/workflow"
)

var (
	eventName    string
	ref          string
	secretsFlags []string

	dryrun = &cobra.Command{
		Use:     "dryrun [workflow-file]",
		Aliases: []string{"dr"},
		Short:   "Analyze a workflow without running it",
		Long: `Dry-run analyzes a GitHub Actions workflow file and shows
what would run based on the current git state and simulated event.

It evaluates all conditions and shows which jobs and steps would
execute, helping you debug your workflows locally.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runDryrun,
	}
)

func init() {
	dryrun.Flags().StringVarP(&eventName, "event", "e", "push", "Event type to simulate (push, pull_request, etc.)")
	dryrun.Flags().StringVarP(&ref, "ref", "r", "", "Git ref to use (defaults to current branch)")
	dryrun.Flags().StringSliceVarP(&secretsFlags, "secret", "s", nil, "Secrets in KEY=VALUE format")
}

func runDryrun(cmd *cobra.Command, args []string) error {
	// Find workflow file
	var workflowPath string
	if len(args) > 0 {
		workflowPath = args[0]
	} else {
		// Find first workflow in .github/workflows
		workflows, err := workflow.FindWorkflows(".")
		if err != nil {
			return fmt.Errorf("finding workflows: %w", err)
		}
		if len(workflows) == 0 {
			return fmt.Errorf("no workflow files found in .github/workflows")
		}
		workflowPath = workflows[0]
		fmt.Fprintf(os.Stderr, "Using workflow: %s\n\n", workflowPath)
	}

	// Parse workflow
	wf, err := workflow.Parse(workflowPath)
	if err != nil {
		return fmt.Errorf("parsing workflow: %w", err)
	}

	// Parse secrets from flags
	secrets := make(map[string]string)
	for _, s := range secretsFlags {
		for i := 0; i < len(s); i++ {
			if s[i] == '=' {
				secrets[s[:i]] = s[i+1:]
				break
			}
		}
	}

	// Build context
	ctx, err := workflow.NewContext(workflow.Options{
		EventName: eventName,
		Ref:       ref,
		Secrets:   secrets,
	})
	if err != nil {
		return fmt.Errorf("building context: %w", err)
	}

	// Analyze
	a := workflow.NewAnalyzer(wf, ctx)
	result := a.Analyze()

	// Render output
	workflow.Render(result)

	return nil
}
