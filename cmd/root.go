package cmd

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/spf13/cobra"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "bitrise-plugin-step",
	Short: "Bitrise CLI Plugin: Step",
	Long:  `Manage, query, and create steps`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(colorstring.Red("ERROR:"), err)
		os.Exit(-1)
	}
}

func init() {

}
