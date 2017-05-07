package create

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/templateutil"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/pkg/errors"
)

const (
	toolkitTypeBash = "bash"
	toolkitTypeGo   = "go"
)

// Inventory ...
type Inventory struct {
	Author         string
	Title          string
	ID             string
	Summary        string
	Description    string
	PrimaryTypeTag string
	//
	ToolkitType string
	//
	Year int
}

// Step ...
func Step() error {
	defaultAuthor := readAuthorFromGitConfig()
	author, err := goinp.AskForStringWithDefault(colorstring.Green("Who are you / who's the author?"), defaultAuthor)
	if err != nil {
		return errors.Wrap(err, "Failed to determine author")
	}

	title, err := goinp.AskForString(colorstring.Green("What's the title / name of the Step?"))
	if err != nil {
		return errors.Wrap(err, "Failed to determine title")
	}

	id := generateIDFromString(title)
	printInfoLine("Generated Step ID (from provided Title):", id)

	summary, err := goinp.AskForString(colorstring.Green("Please provide a summary"))
	if err != nil {
		return errors.Wrap(err, "Failed to determine summary")
	}
	description, err := goinp.AskForString(colorstring.Green("Please provide a description"))
	if err != nil {
		return errors.Wrap(err, "Failed to determine description")
	}

	fmt.Println()
	// available primary categories / type_tags:
	// https://github.com/bitrise-io/bitrise/blob/master/_docs/step-development-guideline.md#step-grouping-convention
	primaryTypeTag, err := goinp.SelectFromStrings(colorstring.Green("What's the primary category of this Step?"), []string{
		"access-control", "artifact-info",
		"installer", "deploy",
		"utility", "dependency", "code-sign",
		"build", "test", "notification",
	})
	if err != nil {
		return errors.Wrap(err, "Failed to determine primary category")
	}

	return createStep(Inventory{
		Author:         author,
		Title:          title,
		ID:             id,
		Summary:        summary,
		Description:    description,
		PrimaryTypeTag: primaryTypeTag,
		//
		ToolkitType: toolkitTypeBash,
		//
		Year: time.Now().Year(),
	})
}

func readAuthorFromGitConfig() string {
	userName, err := command.New("git", "config", "user.name").RunAndReturnTrimmedOutput()
	if err != nil {
		return ""
	}
	return userName
}

func generateIDFromString(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.Map(func(r rune) rune {
		if r < 'a' || r > 'z' {
			return '-'
		}
		return r
	}, s)
	return strings.Trim(s, "-")
}

func printInfoLine(s string, args ...string) {
	parts := append([]string{colorstring.Yellow(s)}, args...)
	fmt.Println(strings.Join(parts, " "))
}
func printSuccessLine(s string, args ...string) {
	parts := append([]string{colorstring.Green(s)}, args...)
	fmt.Println(strings.Join(parts, " "))
}

func createStep(inventory Inventory) error {
	fmt.Println()

	// create directory
	stepDirPth, err := pathutil.AbsPath("bitrise-step-" + inventory.ID)
	if err != nil {
		return errors.Wrap(err, "Failed to get absolute path for step directory")
	}

	printInfoLine("Creating Step directory at:", stepDirPth)
	if exists, err := pathutil.IsPathExists(stepDirPth); err != nil {
		return errors.Wrap(err, "Failed to check whether step dir already exists")
	} else if exists {
		return errors.Errorf("Directory (%s) already exists!", stepDirPth)
	}
	if err := os.Mkdir(stepDirPth, 0755); err != nil {
		return errors.Wrap(err, "Failed to create step directory")
	}

	// save files from templates
	for _, aTemplate := range []struct {
		TemplatePath  string
		FilePath      string
		ToolkitFilter string
	}{
		{
			TemplatePath: "README.md.gotemplate",
			FilePath:     filepath.Join(stepDirPth, "README.md"),
		},
		{
			TemplatePath: "LICENSE.gotemplate",
			FilePath:     filepath.Join(stepDirPth, "LICENSE"),
		},
		{
			TemplatePath: "gitignore.gotemplate",
			FilePath:     filepath.Join(stepDirPth, ".gitignore"),
		},
		{
			TemplatePath: "step.yml.gotemplate",
			FilePath:     filepath.Join(stepDirPth, "step.yml"),
		},
		{
			TemplatePath: "bitrise.yml.gotemplate",
			FilePath:     filepath.Join(stepDirPth, "bitrise.yml"),
		},
		{
			TemplatePath: "bitrise.secrets.yml.gotemplate",
			FilePath:     filepath.Join(stepDirPth, ".bitrise.secrets.yml"),
		},
		// Toolkit: Bash
		{
			TemplatePath:  "bash/step.sh.gotemplate",
			FilePath:      filepath.Join(stepDirPth, "step.sh"),
			ToolkitFilter: toolkitTypeBash,
		},
		// Toolkit: Go
		{
			TemplatePath:  "go/main.go.gotemplate",
			FilePath:      filepath.Join(stepDirPth, "main.go"),
			ToolkitFilter: toolkitTypeGo,
		},
	} {
		if aTemplate.ToolkitFilter != "" && aTemplate.ToolkitFilter != inventory.ToolkitType {
			// skip
			continue
		}

		if err := evaluateTemplateAndWriteToFile(aTemplate.FilePath, aTemplate.TemplatePath, inventory); err != nil {
			return errors.Wrap(err, "Failed to write template into file")
		}
		fmt.Println(" *", colorstring.Green("[OK]"), "created:", aTemplate.FilePath)
	}

	fmt.Println()
	fmt.Println(colorstring.Yellow("Initializing git repository in step directory ..."))
	if err := initGitRepoAtPath(stepDirPth); err != nil {
		return errors.Wrap(err, "Failed to initialize git repository in step directory")
	}

	fmt.Println()
	printSuccessLine("Step is ready!")
	fmt.Println()
	fmt.Println("You can find it at:", stepDirPth)
	fmt.Println()
	fmt.Println("TIP:", colorstring.Yellow("cd"), "into", colorstring.Yellow(stepDirPth), "and run",
		colorstring.Yellow("bitrise run test"), "for a quick test drive!")

	return nil
}

func initGitRepoAtPath(dirPth string) error {
	cmdLog, err := command.New("git", "init").SetDir(dirPth).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "Failed to git init in directory (%s). Output: %s", dirPth, cmdLog)
	}
	return nil
}

func evaluateTemplateAndWriteToFile(filePth, templatePth string, inventory Inventory) error {
	templatesBox, err := rice.FindBox("templates")
	if err != nil {
		return errors.Wrap(err, "Failed to find templates dir/box")
	}

	templateContent, err := templatesBox.String(templatePth)
	if err != nil {
		return errors.Wrapf(err, "Failed to read %s template", templatePth)
	}
	cont, err := templateutil.EvaluateTemplateStringToString(templateContent, inventory, template.FuncMap{})
	if err != nil {
		return errors.Wrapf(err, "Failed to evaluate template %s", templatePth)
	}
	if err := fileutil.WriteStringToFile(filePth, cont); err != nil {
		return errors.Wrapf(err, "Failed to write evaluated template into file (%s)", filePth)
	}
	return nil
}
