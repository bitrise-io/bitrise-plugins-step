package cmd

import (
	"errors"
	"fmt"

	"github.com/bitrise-io/bitrise/output"
	"github.com/bitrise-io/bitrise/tools"
	"github.com/spf13/cobra"
)

var (
	collection = ""
	format     = ""
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List of available steps",
	Long:  "List of available steps",
	RunE: func(cmd *cobra.Command, args []string) error {
		return printStepList()
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&collection, "collection", "c", "", "Collection of step.")
	listCmd.Flags().StringVar(&format, "format", "", "Output format. Accepted: raw, json.")
}

func printStepList() error {
	if collection == "" {
		return errors.New("no collection defined")
	}
	switch format {
	case "", output.FormatRaw:
		out, err := tools.StepmanRawStepList(collection)
		if err != nil {
			return fmt.Errorf("failed to print step list, err: %s", err)
		}
		fmt.Println("Step list:")
		fmt.Println(out)
	case output.FormatJSON:
		out, err := tools.StepmanJSONStepList(collection)
		if err != nil {
			return fmt.Errorf("failed to print step list, err: %s", err)
		}
		fmt.Println(out)
	default:
		return fmt.Errorf("invalid format: %s", format)
	}
	return nil
}
