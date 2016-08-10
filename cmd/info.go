package cmd

import (
	"errors"
	"fmt"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"

	"github.com/bitrise-core/bitrise-plugins-step/stepmanutil"
	"github.com/bitrise-core/bitrise-plugins-step/utils"
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

		collectionID := "https://github.com/bitrise-io/bitrise-steplib.git"
		_, stepVersion, err := stepmanutil.ReadStepVersionInfo(collectionID, stepID, stepVersion)

		if err != nil {
			return fmt.Errorf("Failed to get step info: %s", err)
		}

		// Check if step exist in collection
		collection, err := stepmanutil.ReadStepCollectionModel(collectionID)
		if err != nil {
			return fmt.Errorf("Failed to read steps spec (spec.json), err: %s", err)
		}

		step, stepFound := collection.GetStep(stepID, stepVersion)
		if !stepFound {
			if stepVersion == "" {
				return fmt.Errorf("Collection doesn't contain any version of step (id:%s)", stepID)
			}
			return fmt.Errorf("Collection doesn't contain step (id:%s) (version:%s)", stepID, stepVersion)
		}

		latestStepVersion, err := collection.GetLatestStepVersion(stepID)
		if err != nil {
			return fmt.Errorf("Failed to get latest version of step (id:%s)", stepID)
		}

		if stepVersion == "" {
			stepVersion = latestStepVersion
		}

		inputs, err := getEnvInfos(step.Inputs)
		if err != nil {
			return fmt.Errorf("Failed to get step (id:%s) input infos, err: %s", stepID, err)
		}

		outputs, err := getEnvInfos(step.Outputs)
		if err != nil {
			return fmt.Errorf("Failed to get step (id:%s) output infos, err: %s", stepID, err)
		}

		stepInfo := models.StepInfoModel{
			ID:          stepID,
			Version:     stepVersion,
			Latest:      latestStepVersion,
			Description: *step.Description,
			StepLib:     collectionID,
			Source:      *step.SourceCodeURL,
			Inputs:      inputs,
			Outputs:     outputs,
		}

		route, found := stepman.ReadRoute(collectionID)
		if !found {
			return fmt.Errorf("No route found for collection: %s", collectionID)
		}
		globalStepInfoPth := stepman.GetStepGlobalInfoPath(route, stepID)
		if globalStepInfoPth != "" {
			globalInfo, found, err := stepman.ParseGlobalStepInfoYML(globalStepInfoPth)
			if err != nil {
				return fmt.Errorf("Failed to get step (path:%s) output infos, err: %s", globalStepInfoPth, err)
			}

			if found {
				stepInfo.GlobalInfo = globalInfo
			}
		}

		printStepVersionInfoOutput(stepInfo)

		return nil
	},
}

func init() {
	RootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringVarP(&stepVersion, "version", "v", "", "Version - if not specified will print info about the latest version")
}

func getEnvInfos(envs []envmanModels.EnvironmentItemModel) ([]models.EnvInfoModel, error) {
	envInfos := []models.EnvInfoModel{}
	for _, env := range envs {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return []models.EnvInfoModel{}, err
		}

		options, err := env.GetOptions()
		if err != nil {
			return []models.EnvInfoModel{}, err
		}

		envInfo := models.EnvInfoModel{
			Key:          key,
			Description:  *options.Description,
			ValueOptions: options.ValueOptions,
			DefaultValue: value,
			IsExpand:     *options.IsExpand,
		}
		envInfos = append(envInfos, envInfo)
	}
	return envInfos, nil
}

func printStepVersionInfoOutput(stepVersionInfo models.StepInfoModel) error {
	fmt.Println(colorstring.Green(stepVersionInfo.ID) + " (" + stepVersionInfo.Version + ")")
	fmt.Println()
	fmt.Println(colorstring.Yellow("Description") + ": ")
	fmt.Println(utils.IndentTextWithMaxLength(stepVersionInfo.Description, "", 80))
	if len(stepVersionInfo.Inputs) > 0 {
		fmt.Println()
		fmt.Println(colorstring.Blue("=== Inputs =========="))
		for _, input := range stepVersionInfo.Inputs {
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
