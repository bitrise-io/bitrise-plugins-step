package cmd

import (
	"github.com/bitrise-io/bitrise-plugins-step/create"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Step",
	Long:  `Answer a couple of questions and have a fully working step in seconds!`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return create.Step()
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
}
