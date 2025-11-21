package cmds

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/telton/rehearse/workflow"
)

var (
	eventName        string
	ref              string
	eventPayloadFile string

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

			wf, err := workflow.New(f)
			cobra.CheckErr(err)

			var eventPayload map[string]any
			if eventPayloadFile != "" {
				evFile, err := os.Open(eventPayloadFile)
				cobra.CheckErr(err)

				err = json.NewDecoder(evFile).Decode(&eventPayload)
				cobra.CheckErr(err)
			}

			wfCtx, err := workflow.NewContext(eventName, ref, eventPayload)
			cobra.CheckErr(err)

			wf.DryRun(wfCtx)
		},
	}
)

func init() {
	dryrun.PersistentFlags().StringVar(&eventName, "event-name", "push", "The GitHub event to simulate")
	dryrun.PersistentFlags().StringVar(&ref, "ref", "", "The git ref to use")
	dryrun.PersistentFlags().StringVar(&eventPayloadFile, "event-payload-file", "", "The filepath to the event payload")
}
