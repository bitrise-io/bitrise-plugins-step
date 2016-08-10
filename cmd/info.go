package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bitrise-core/bitrise-plugins-step/stepman"
	"github.com/bitrise-core/bitrise-plugins-step/utils"
	"github.com/bitrise-io/go-utils/cmdex"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/spf13/cobra"
)

var (
	stepVersion = ""
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print info about a step",
	Long:  `Print info about a step`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("No step ID specified as a parameter")
		}
		if len(args) > 1 {
			return fmt.Errorf("More than one step ID specified: %s", args)
		}
		stepID := args[0]

		out, err := cmdex.RunCommandAndReturnStdout("stepman",
			"step-info",
			"--collection", "https://github.com/bitrise-io/bitrise-steplib.git",
			"--format", "json",
			"--id", stepID,
		)

		if err != nil {
			return fmt.Errorf("Failed to get step info from stepman: %s", err)
		}

		if err := printStepVersionInfoOutput(out, stepVersion); err != nil {
			return fmt.Errorf("Failed to format stepman output: %s", err)
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringVarP(&stepVersion, "version", "v", "", "Version - if not specified will print info about the latest version")
}

func printStepVersionInfoOutput(jsonOutput string, versionToPrint string) error {
	var stepVerInfo stepman.StepVersionModel
	if err := json.Unmarshal([]byte(jsonOutput), &stepVerInfo); err != nil {
		return fmt.Errorf("Failed to parse JSON: %s", err)
	}
	fmt.Println(colorstring.Green(stepVerInfo.ID) + " (" + stepVerInfo.Version + ")")
	fmt.Println()
	fmt.Println(colorstring.Yellow("Description") + ": ")
	fmt.Println(utils.IndentTextWithMaxLength(stepVerInfo.Description, "", 80))
	if len(stepVerInfo.Inputs) > 0 {
		fmt.Println()
		fmt.Println(colorstring.Blue("Inputs:"))
		for _, input := range stepVerInfo.Inputs {
			fmt.Println()
			fmt.Println(colorstring.Green(input.Key) + ":")

			// DefaultValue string   `json:"default_value"`
			// ValueOptions []string `json:"value_options"`
			// IsExpand     bool     `json:"is_expand"`
			fmt.Println(colorstring.Yellow("  Description") + ":")
			fmt.Println(utils.IndentTextWithMaxLength(input.Description, "  ", 80))
		}
	}
	fmt.Println()
	return nil
}
