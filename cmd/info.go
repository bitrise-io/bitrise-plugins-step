package cmd

import (
	"encoding/json"
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

		// if err := printStepVersionInfoOutput(stepVersionInfo, stepID, stepVersion); err != nil {
		// 	return fmt.Errorf("Failed to format stepman output: %s", err)
		// }

		//
		// StepLib step info

		// Input validation
		id := stepID
		collectionURI := collectionID
		version := stepVersion

		if id == "" {
			return errors.New("Missing required input: step id")
		}

		// Check if step exist in collection
		collection, err := stepman.ReadStepSpec(collectionURI)
		if err != nil {
			return fmt.Errorf("Failed to read steps spec (spec.json), err: %s", err)
		}

		step, stepFound := collection.GetStep(id, version)
		if !stepFound {
			if version == "" {
				return fmt.Errorf("Collection doesn't contain any version of step (id:%s)", id)
			}
			return fmt.Errorf("Collection doesn't contain step (id:%s) (version:%s)", id, version)
		}

		latest, err := collection.GetLatestStepVersion(id)
		if err != nil {
			return fmt.Errorf("Failed to get latest version of step (id:%s)", id)
		}

		if version == "" {
			version = latest
		}

		inputs, err := getEnvInfos(step.Inputs)
		if err != nil {
			return fmt.Errorf("Failed to get step (id:%s) input infos, err: %s", id, err)
		}

		outputs, err := getEnvInfos(step.Outputs)
		if err != nil {
			return fmt.Errorf("Failed to get step (id:%s) output infos, err: %s", id, err)
		}

		stepInfo := models.StepInfoModel{
			ID:          id,
			Version:     version,
			Latest:      latest,
			Description: *step.Description,
			StepLib:     collectionURI,
			Source:      *step.SourceCodeURL,
			Inputs:      inputs,
			Outputs:     outputs,
		}

		route, found := stepman.ReadRoute(collectionURI)
		if !found {
			return fmt.Errorf("No route found for collection: %s", collectionURI)
		}
		globalStepInfoPth := stepman.GetStepGlobalInfoPath(route, id)
		if globalStepInfoPth != "" {
			globalInfo, found, err := stepman.ParseGlobalStepInfoYML(globalStepInfoPth)
			if err != nil {
				return fmt.Errorf("Failed to get step (path:%s) output infos, err: %s", globalStepInfoPth, err)
			}

			if found {
				stepInfo.GlobalInfo = globalInfo
			}
		}

		// printRawStepInfo(stepInfo, false, false)
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

func printRawEnvInfo(env models.EnvInfoModel) {
	if env.DefaultValue != "" {
		fmt.Printf("- %s: %s\n", colorstring.Green(env.Key), env.DefaultValue)
	} else {
		fmt.Printf("- %s\n", colorstring.Green(env.Key))
	}

	fmt.Printf("  %s: %v\n", colorstring.Green("is expand"), env.IsExpand)

	if len(env.ValueOptions) > 0 {
		fmt.Printf("  %s:\n", colorstring.Green("value options"))
		for _, option := range env.ValueOptions {
			fmt.Printf("  - %s\n", option)
		}
	}

	if env.Description != "" {
		fmt.Printf("  %s:\n", colorstring.Green("description"))
		fmt.Printf("  %s\n", env.Description)
	}
}

func printRawStepInfo(stepInfo models.StepInfoModel, isShort, isLocal bool) {
	if isLocal {
		fmt.Println(colorstring.Bluef("Local step info, yml path (%s):", stepInfo.StepLib))
	} else {
		fmt.Println(colorstring.Bluef("Step info in StepLib (%s):", stepInfo.StepLib))
	}

	if stepInfo.GlobalInfo.RemovalDate != "" {
		fmt.Println("")
		fmt.Println(colorstring.Red("This step is deprecated!"))
		fmt.Printf("%s %s\n", colorstring.Red("removal date:"), stepInfo.GlobalInfo.RemovalDate)

		if stepInfo.GlobalInfo.DeprecateNotes != "" {
			fmt.Printf("%s\n%s\n", colorstring.Red("deprecate notes:"), stepInfo.GlobalInfo.DeprecateNotes)
		}
	}

	if stepInfo.ID != "" {
		fmt.Printf("%s: %s\n", colorstring.Blue("ID"), stepInfo.ID)
	}
	if stepInfo.Version != "" {
		fmt.Printf("%s: %s\n", colorstring.Blue("version"), stepInfo.Version)
	}
	if stepInfo.Latest != "" {
		fmt.Printf("%s: %s\n", colorstring.Blue("latest"), stepInfo.Latest)
	}

	if !isShort {
		fmt.Printf("%s: %s\n", colorstring.Blue("source"), stepInfo.Source)
		fmt.Printf("%s:\n", colorstring.Blue("description"))
		fmt.Printf("%s\n", stepInfo.Description)
		fmt.Println()

		if len(stepInfo.Inputs) > 0 {
			fmt.Printf("%s:\n", colorstring.Blue("inputs"))
			for _, input := range stepInfo.Inputs {
				printRawEnvInfo(input)
			}
		}

		if len(stepInfo.Outputs) > 0 {
			if len(stepInfo.Inputs) > 0 {
				fmt.Println()
			}
			fmt.Printf("%s:\n", colorstring.Blue("outputs"))
			for _, output := range stepInfo.Outputs {
				printRawEnvInfo(output)
			}
		}
	}

	fmt.Println()
	fmt.Println()
}

func printJSONStepInfo(stepInfo models.StepInfoModel, isShort bool) error {
	bytes, err := json.Marshal(stepInfo)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))
	return nil
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
