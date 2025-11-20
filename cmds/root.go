package cmds

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "rehearse",
	Short: "Practice before the real thing",
	Long:  `Rehearse is a CLI to debug and step through your GitHub Action workflows.`,
}

func Execute() error {
	return rootCmd.Execute()
}
