package cmd

import (
	"fmt"
	"runtime"

	"github.com/bitrise-core/bitrise-plugins-step/version"
	"github.com/spf13/cobra"
)

var (
	isFullVersionPrint = false
)

// VersionOutputModel ...
type VersionOutputModel struct {
	Version     string `json:"version"`
	OS          string `json:"os"`
	GO          string `json:"go"`
	BuildNumber string `json:"build_number"`
	Commit      string `json:"commit"`
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Long: `Prints the version of the plugin.

Use the --full flag to print extended version infos.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		versionOutput := VersionOutputModel{
			Version: version.VERSION,
		}

		if isFullVersionPrint {
			versionOutput.BuildNumber = version.BuildNumber
			versionOutput.Commit = version.Commit
			versionOutput.OS = fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)
			versionOutput.GO = runtime.Version()
		}

		if isFullVersionPrint {
			versionStr := fmt.Sprintf(`version: %s
os: %s
go: %s
build number: %s
commit: %s
`, versionOutput.Version, versionOutput.OS, versionOutput.GO, versionOutput.BuildNumber, versionOutput.Commit)
			fmt.Println(versionStr)
		} else {
			versionStr := fmt.Sprintf("%s", versionOutput.Version)
			fmt.Println(versionStr)
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&isFullVersionPrint, "full", false, "Full / extended version infos")
}
