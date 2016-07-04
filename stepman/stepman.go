package stepman

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

const (
	// hardcoded, while stepman does not have this feature
	stepmanRoutesPath = "~/.stepman/routing.json"
)

// StepInputModel ...
type StepInputModel struct {
	Key          string   `json:"key"`
	Description  string   `json:"description"`
	DefaultValue string   `json:"default_value"`
	ValueOptions []string `json:"value_options"`
	IsExpand     bool     `json:"is_expand"`
}

// StepVersionModel ...
type StepVersionModel struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Inputs      []StepInputModel `json:"inputs"`
}

// StepInfoModel ...
type StepInfoModel struct {
	LatestVersion string
	StepVersions  map[string]StepVersionModel `json:"versions"`
}

// SpecJSONModel ...
type SpecJSONModel struct {
	Steps map[string]StepInfoModel `json:"steps"`
}

// ReadStepInfo ...
func ReadStepInfo(collectionID, stepID, stepVersion string) (StepVersionModel, error) {
	specJSONPath, err := specJSONPathOfCollection(collectionID)
	if err != nil {
		return StepVersionModel{}, fmt.Errorf("Failed to get spec json path: %s", err)
	}

	file, err := os.Open(specJSONPath)
	if err != nil {
		return StepVersionModel{}, fmt.Errorf("Failed to open spec json: %s", err)
	}
	var spec SpecJSONModel
	if err := json.NewDecoder(file).Decode(&spec); err != nil {
		return StepVersionModel{}, fmt.Errorf("Failed to parse spec json: %s", err)
	}

	return StepVersionModel{}, nil
}

func specJSONPathOfCollection(collectionID string) (string, error) {
	routesAbsPath, err := pathutil.AbsPath(stepmanRoutesPath)
	if err != nil {
		return "", fmt.Errorf("Failed to get absolut path for stepman routing file: %s", err)
	}
	bytes, err := fileutil.ReadBytesFromFile(routesAbsPath)
	if err != nil {
		return "", fmt.Errorf("Failed to read content of routing file: %s", err)
	}
	var routeMap map[string]string
	if err := json.Unmarshal(bytes, &routeMap); err != nil {
		return "", fmt.Errorf("Failed to parse content of routing file: %s", err)
	}

	val, isFound := routeMap[collectionID]
	if !isFound {
		return "", fmt.Errorf("Specified collection (%s) not found in routing", collectionID)
	}

	specPath := fmt.Sprintf("~/.stepman/step_collections/%s/spec/spec.json", val)
	absSpecJSONPath, err := pathutil.AbsPath(specPath)
	if err != nil {
		return "", fmt.Errorf("Failed to get absolute path of spec.json")
	}

	return absSpecJSONPath, nil
}
