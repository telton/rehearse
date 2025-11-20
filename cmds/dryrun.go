package cmds

import (
	"errors"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/telton/rehearse/parser"
)

var (
	event string

	dryrun = &cobra.Command{
		Use:     "dryrun [workflow file]",
		Aliases: []string{"dr"},
		Short:   "Run a workflow in dryrun mode",
		Long:    ``,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			var comps []string

			return comps, cobra.ShellCompDirectiveDefault
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				cobra.CheckErr(errors.New("dryrun needs a workflow file"))
			}

			f, err := os.Open(args[0])
			cobra.CheckErr(err)
			defer f.Close()

			wf, err := parser.NewWorkflow(f)
			cobra.CheckErr(err)

			log.Infof("Starting workflow: %s", wf.Name)

			if event == "pull_request" && wf.On.PullRequest != nil {
				log.Infof("Pull request event triggered")
				if len(wf.On.PullRequest.Paths) > 0 {
					log.Info("Paths:")
				}
				for _, p := range wf.On.PullRequest.Paths {
					log.Printf("\t %s", p)
				}
			}
			if event == "push" && wf.On.Push != nil {
				log.Infof("Push event triggered")
				if len(wf.On.Push.Branches) > 0 {
					log.Info("Branches:")
				}
				for _, b := range wf.On.Push.Branches {
					log.Printf("\t %s", b)
				}
				if len(wf.On.Push.Paths) > 0 {
					log.Info("Paths:")
				}
				for _, p := range wf.On.Push.Paths {
					log.Printf("\t %s", p)
				}
			}
			// TODO: add workflow_dispatch

			log.Info("Workflow run complete")
		},
	}
)

func init() {
	dryrun.PersistentFlags().StringVar(&event, "event", "push", "The GitHub event to simulate")
}
