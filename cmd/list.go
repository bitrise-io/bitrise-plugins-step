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
		return errors.New("No collection defined")
	}
	if format == "" || format == output.FormatRaw {
		out, err := tools.StepmanRawStepList(collection)
		if err != nil {
			return fmt.Errorf("Failed to print step list, err: %s", err)
		}
		if out != "" {
			fmt.Printf("Step list:\n%s\n", out)
		}
	} else if format == output.FormatJSON {
		out, err := tools.StepmanJSONStepList(collection)
		if err != nil {
			return fmt.Errorf("Failed to print step list, err: %s", err)
		}
		fmt.Println(out)
	} else {
		return fmt.Errorf("Invalid format: %s", format)
	}
	return nil
}
