package cmd

import (
	"errors"
	"fmt"

	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/pointers"
	"github.com/bitrise-io/stepman/models"
	"github.com/bitrise-io/stepman/stepman"

	"github.com/bitrise-io/bitrise-plugins-step/stepmanutil"
	"github.com/bitrise-io/bitrise-plugins-step/utils"
	"github.com/spf13/cobra"
)

var (
	stepVersion  = ""
	stepYMLPath  = ""
	outputFormat = ""
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print info about a step",
	Long:  `Print info about a step`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// from step.yml
		if len(stepYMLPath) != 0 {
			return printStepInfoFromStepYML(stepYMLPath)
		}

		// by Step ID
		if len(args) < 1 {
			return errors.New("No step ID specified as a parameter")
		}
		if len(args) > 1 {
			return fmt.Errorf("More than one step ID specified: %s", args)
		}
		stepID := args[0]

		return printStepInfoFromLibrary(stepID)
	},
}

func printStepInfoFromStepYML(ymlPth string) error {
	step, err := stepman.ParseStepDefinition(ymlPth, false)
	if err != nil {
		return fmt.Errorf("Failed to parse step.yml (path: %s), error: %s", ymlPth, err)
	}

	// inputs, err := getEnvInfos(step.Inputs)
	// if err != nil {
	// 	return fmt.Errorf("Failed to get step input infos, err: %s", err)
	// }

	// outputs, err := getEnvInfos(step.Outputs)
	// if err != nil {
	// 	return fmt.Errorf("Failed to get step output infos, err: %s", err)
	// }

	stepInfo := models.StepInfoModel{
		Library:       "",
		ID:            "step.yml:" + ymlPth,
		Version:       "",
		LatestVersion: "",
		Step: models.StepModel{
			Description:   step.Description,
			SourceCodeURL: step.SourceCodeURL,
			SupportURL:    step.SupportURL,
			Inputs:        step.Inputs,
			Outputs:       step.Outputs,
		},
	}

	return printStepVersionInfoOutput(stepInfo)
}

func printStepInfoFromLibrary(stepID string) error {
	collectionID := "https://github.com/bitrise-io/bitrise-steplib.git"
	_, stepVersion, err := stepmanutil.ReadStepVersionInfo(collectionID, stepID, stepVersion)

	if err != nil {
		return fmt.Errorf("Failed to get step info: %s", err)
	}

	step, err := stepman.ReadStepVersionInfo(collectionID, stepID, stepVersion)
	if err != nil {
		return fmt.Errorf("Failed to read step version info: %s", err)
	}

	collection, err := stepman.ReadStepSpec(collectionID)
	if err != nil {
		return fmt.Errorf("Failed to read step lib (%s), error: %s", collectionID, err)
	}

	latestStepVersion, err := collection.GetLatestStepVersion(stepID)
	if err != nil {
		return fmt.Errorf("Failed to get latest version of step (id:%s)", stepID)
	}

	if stepVersion == "" {
		stepVersion = latestStepVersion
	}

	// inputs, err := getEnvInfos(step.Inputs)
	// if err != nil {
	// 	return fmt.Errorf("Failed to get step (id:%s) input infos, err: %s", stepID, err)
	// }

	// outputs, err := getEnvInfos(step.Outputs)
	// if err != nil {
	// 	return fmt.Errorf("Failed to get step (id:%s) output infos, err: %s", stepID, err)
	// }

	stepInfo := models.StepInfoModel{
		Library:       collectionID,
		ID:            stepID,
		Version:       stepVersion,
		LatestVersion: latestStepVersion,
		Step: models.StepModel{
			Description:   step.Step.Description,
			SourceCodeURL: step.Step.SourceCodeURL,
			SupportURL:    step.Step.SupportURL,
			Inputs:        step.Step.Inputs,
			Outputs:       step.Step.Outputs,
		},
	}

	route, found := stepman.ReadRoute(collectionID)
	if !found {
		return fmt.Errorf("No route found for collection: %s", collectionID)
	}
	globalStepInfoPth := stepman.GetStepGlobalInfoPath(route, stepID)
	if globalStepInfoPth != "" {
		globalInfo, found, err := stepman.ParseStepGroupInfoModel(globalStepInfoPth)
		if err != nil {
			return fmt.Errorf("Failed to get step (path:%s) output infos, err: %s", globalStepInfoPth, err)
		}

		if found {
			stepInfo.GroupInfo = globalInfo
		}
	}

	return printStepVersionInfoOutput(stepInfo)
}

func init() {
	RootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringVarP(&stepVersion, "version", "v", "", "Version - if not specified will print info about the latest version")
	infoCmd.Flags().StringVar(&stepYMLPath, "step-yml", "", "step.yml - if specified infos will be printed from the specified step.yml, not from a library")
	infoCmd.Flags().StringVar(&outputFormat, "output-format", "", `Output format. Default is "rich command line", but can also be "markdown", to generate a standard markdown output instead.`)
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
	isMarkdown := (outputFormat == "markdown")

	// Step ID, collection, version, ...
	if isMarkdown {
		fmt.Println("# " + stepVersionInfo.ID)
		fmt.Println()
		fmt.Println("- version: " + stepVersionInfo.Version)
		fmt.Println("- collection: " + stepVersionInfo.Library)
	} else {
		fmt.Println(colorstring.Green(stepVersionInfo.ID) + "  @" + stepVersionInfo.Version + "  [" + stepVersionInfo.Library + "]")
		fmt.Println()
	}
	// base infos like support & source URL
	if isMarkdown {
		fmt.Println()
		fmt.Println("# Base Infos")
		fmt.Println()
		fmt.Println("- Support URL: " + pointers.String(stepVersionInfo.Step.SupportURL))
		fmt.Println("- Source URL: " + pointers.String(stepVersionInfo.Step.SourceCodeURL))
	} else {
		fmt.Println(colorstring.Yellow("Support") + ": " + pointers.String(stepVersionInfo.Step.SupportURL))
		fmt.Println(colorstring.Yellow("Source") + ": " + pointers.String(stepVersionInfo.Step.SourceCodeURL))
		fmt.Println()
	}
	// description
	if isMarkdown {
		fmt.Println()
		fmt.Println("# Description")
		fmt.Println()
		fmt.Println(pointers.String(stepVersionInfo.Step.Description))
	} else {
		fmt.Println(colorstring.Yellow("Description") + ": ")
		fmt.Println(utils.IndentTextWithMaxLength(pointers.String(stepVersionInfo.Step.Description), "", 80))
	}

	// inputs
	if len(stepVersionInfo.Step.Inputs) > 0 {
		if isMarkdown {
			fmt.Println()
			fmt.Println("# Inputs")
		} else {
			fmt.Println()
			fmt.Println(colorstring.Blue("=== Inputs =========="))
		}

		inputs, err := getEnvInfos(stepVersionInfo.Step.Inputs)
		if err != nil {
			return fmt.Errorf("Failed to get step input infos, err: %s", err)
		}

		for _, input := range inputs {
			if isMarkdown {
				fmt.Println()
				fmt.Println("## `" + input.Key + "`")
				fmt.Println()
				fmt.Println("### _Description_")
				fmt.Println()
				fmt.Println(input.Description)
			} else {
				fmt.Println()
				fmt.Println(colorstring.Green(input.Key) + ":")

				// DefaultValue string   `json:"default_value"`
				// ValueOptions []string `json:"value_options"`
				// IsExpand     bool     `json:"is_expand"`
				fmt.Println(colorstring.Yellow("  Description") + ":")
				fmt.Println(utils.IndentTextWithMaxLength(input.Description, "  ", 80))
			}
		}
	}
	fmt.Println()
	return nil
}
