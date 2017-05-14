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

// GoToolkitInventoryModel ...
type GoToolkitInventoryModel struct {
	// PackageID: e.g.: github.com/bitrise-io/bitrise
	PackageID string
}

// InventoryModel ...
type InventoryModel struct {
	Author         string
	Title          string
	ID             string
	Summary        string
	Description    string
	PrimaryTypeTag string
	//
	ToolkitType        string
	GoToolkitInventory GoToolkitInventoryModel
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

	// available primary categories / type_tags:
	// https://github.com/bitrise-io/bitrise/blob/master/_docs/step-development-guideline.md#step-grouping-convention
	fmt.Println()
	primaryTypeTag, err := goinp.SelectFromStrings(colorstring.Green("What's the primary category of this Step?"), []string{
		"access-control", "artifact-info",
		"installer", "deploy",
		"utility", "dependency", "code-sign",
		"build", "test", "notification",
	})
	if err != nil {
		return errors.Wrap(err, "Failed to determine primary category")
	}

	fmt.Println()
	fmt.Println("Toolkit: the entry/base language of the Step.")
	fmt.Println("Our recommendation is to use Bash for very simple Steps")
	fmt.Println(" and for more complex ones use another language, one which we have toolkit support for.")
	fmt.Println("If you're just getting started with Step development our suggestion is to select Bash,")
	fmt.Println(" as that's the easiest option. It's possible to convert the step later, if needed.")
	fmt.Println("Note: Of course even if you select e.g. Bash as the entry language, you can run other scripts from there,")
	fmt.Println(" so it's possible to write the majority of the step's code in e.g. Ruby,")
	fmt.Println(" and have an entry Bash script which does nothing else except running the Ruby script.")
	toolkitType, err := goinp.SelectFromStrings(colorstring.Green("Which toolkit (language) would you like to use?"), []string{
		toolkitTypeBash, toolkitTypeGo,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to determine the toolkit")
	}

	goInv := GoToolkitInventoryModel{}
	if toolkitType == toolkitTypeGo {
		fmt.Println()
		fmt.Println("Go programs require a Go package ID, in order to work well with the standard Go tools.")
		fmt.Println("The package ID looks like this usually: SOURCE-CODE-HOSTING-SERVICE/user/package-name")
		fmt.Println(" Example: github.com/bitrise-io/bitrise")
		fmt.Println("If you (plan to) use GitHub for hosting this step's source code,")
		fmt.Println("the suggested package name for this step is:",
			colorstring.Yellow("github.com/YOUR-GITHUB-USERNAME/"+stepDirAndRepoNameFromID(id)))
		goPkgID, err := goinp.AskForString(colorstring.Green("What should be the Go package ID?"))
		if err != nil {
			return errors.Wrap(err, "Failed to determine the package ID")
		}
		goInv.PackageID = goPkgID
	}

	return createStep(InventoryModel{
		Author:         author,
		Title:          title,
		ID:             id,
		Summary:        summary,
		Description:    description,
		PrimaryTypeTag: primaryTypeTag,
		//
		ToolkitType:        toolkitType,
		GoToolkitInventory: goInv,
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
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
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

func stepDirAndRepoNameFromID(stepID string) string {
	return "bitrise-step-" + stepID
}

func createStep(inventory InventoryModel) error {
	fmt.Println()

	// create directory
	stepDirAbsPth := ""
	if inventory.ToolkitType == toolkitTypeBash {
		baseDirPath := stepDirAndRepoNameFromID(inventory.ID)
		absPth, err := pathutil.AbsPath(baseDirPath)
		if err != nil {
			return errors.Wrapf(err, "Failed to get absolute path for step directory (%s)", baseDirPath)
		}
		stepDirAbsPth = absPth
	} else if inventory.ToolkitType == toolkitTypeGo {
		gopath := os.Getenv("GOPATH")
		if len(gopath) < 1 {
			// no GOPATH env set - use "${HOME}/go", which is the default GOPATH since Go 1.8
			gopath = filepath.Join(pathutil.UserHomeDir(), "go")
		}
		baseDirPath := filepath.Join(gopath, "src", inventory.GoToolkitInventory.PackageID)

		absPth, err := pathutil.AbsPath(baseDirPath)
		if err != nil {
			return errors.Wrapf(err, "Failed to get absolute path for step directory (%s)", baseDirPath)
		}
		stepDirAbsPth = absPth
	} else {
		return errors.Errorf("Invalid Toolkit Type: %s", inventory.ToolkitType)
	}

	printInfoLine("Creating Step directory at:", stepDirAbsPth)
	if exists, err := pathutil.IsPathExists(stepDirAbsPth); err != nil {
		return errors.Wrap(err, "Failed to check whether step dir already exists")
	} else if exists {
		return errors.Errorf("Directory (%s) already exists!", stepDirAbsPth)
	}
	if err := os.MkdirAll(stepDirAbsPth, 0755); err != nil {
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
			FilePath:     filepath.Join(stepDirAbsPth, "README.md"),
		},
		{
			TemplatePath: "LICENSE.gotemplate",
			FilePath:     filepath.Join(stepDirAbsPth, "LICENSE"),
		},
		{
			TemplatePath: "gitignore.gotemplate",
			FilePath:     filepath.Join(stepDirAbsPth, ".gitignore"),
		},
		{
			TemplatePath: "step.yml.gotemplate",
			FilePath:     filepath.Join(stepDirAbsPth, "step.yml"),
		},
		{
			TemplatePath: "bitrise.yml.gotemplate",
			FilePath:     filepath.Join(stepDirAbsPth, "bitrise.yml"),
		},
		{
			TemplatePath: "bitrise.secrets.yml.gotemplate",
			FilePath:     filepath.Join(stepDirAbsPth, ".bitrise.secrets.yml"),
		},
		// Toolkit: Bash
		{
			TemplatePath:  "bash/step.sh.gotemplate",
			FilePath:      filepath.Join(stepDirAbsPth, "step.sh"),
			ToolkitFilter: toolkitTypeBash,
		},
		// Toolkit: Go
		{
			TemplatePath:  "go/main.go.gotemplate",
			FilePath:      filepath.Join(stepDirAbsPth, "main.go"),
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
	if err := initGitRepoAtPath(stepDirAbsPth); err != nil {
		return errors.Wrap(err, "Failed to initialize git repository in step directory")
	}

	fmt.Println()
	printSuccessLine("Step is ready!")
	fmt.Println()
	fmt.Println("You can find it at:", stepDirAbsPth)
	fmt.Println()
	fmt.Println("TIP:", colorstring.Yellow("cd"), "into", colorstring.Yellow(stepDirAbsPth), "and run",
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

func evaluateTemplate(templatePth string, inventory InventoryModel) (string, error) {
	templatesBox, err := rice.FindBox("templates")
	if err != nil {
		return "", errors.Wrap(err, "Failed to find templates dir/box")
	}

	templateContent, err := templatesBox.String(templatePth)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to read %s template", templatePth)
	}
	evaluatedContent, err := templateutil.EvaluateTemplateStringToString(templateContent, inventory, template.FuncMap{})
	if err != nil {
		return "", errors.Wrapf(err, "Failed to evaluate template %s", templatePth)
	}
	return evaluatedContent, nil
}

func evaluateTemplateAndWriteToFile(filePth, templatePth string, inventory InventoryModel) error {
	evaluatedContent, err := evaluateTemplate(templatePth, inventory)
	if err != nil {
		return errors.Wrap(err, "Failed to evaluate template")
	}

	if err := fileutil.WriteStringToFile(filePth, evaluatedContent); err != nil {
		return errors.Wrapf(err, "Failed to write evaluated template into file (%s)", filePth)
	}
	return nil
}
