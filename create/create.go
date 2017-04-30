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
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/templateutil"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/pkg/errors"
)

// Inventory ...
type Inventory struct {
	Author      string
	Title       string
	ID          string
	Summary     string
	Description string
	//
	Year int
}

// Step ...
func Step() error {
	author, err := goinp.AskForString("Who's the author?")
	if err != nil {
		return errors.Wrap(err, "Failed to determine author")
	}
	title, err := goinp.AskForString("What's the title / name of the Step?")
	if err != nil {
		return errors.Wrap(err, "Failed to determine title")
	}

	id := generateIDFromString(title)
	printInfoLine("Generated Step ID (from provided Title):", id)

	summary, err := goinp.AskForString("Please provide a summary")
	if err != nil {
		return errors.Wrap(err, "Failed to determine summary")
	}
	description, err := goinp.AskForString("Please provide a description")
	if err != nil {
		return errors.Wrap(err, "Failed to determine description")
	}

	return createStep(Inventory{
		Author:      author,
		Title:       title,
		ID:          id,
		Summary:     summary,
		Description: description,
		//
		Year: time.Now().Year(),
	})
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

func createStep(inventory Inventory) error {
	// create directory
	stepDirPth, err := pathutil.AbsPath(inventory.ID)
	if err != nil {
		return errors.Wrap(err, "Failed to get absolute path for step directory")
	}

	printInfoLine("Creating Step directory:", stepDirPth)
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
		TemplatePath string
		FilePath     string
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
			TemplatePath: "step.sh.gotemplate",
			FilePath:     filepath.Join(stepDirPth, "step.sh"),
		},
		{
			TemplatePath: "bitrise.yml.gotemplate",
			FilePath:     filepath.Join(stepDirPth, "bitrise.yml"),
		},
		{
			TemplatePath: "bitrise.secrets.yml.gotemplate",
			FilePath:     filepath.Join(stepDirPth, ".bitrise.secrets.yml"),
		},
	} {
		if err := evaluateTemplateAndWriteToFile(aTemplate.FilePath, aTemplate.TemplatePath, inventory); err != nil {
			return errors.Wrap(err, "Failed to write template into file")
		}
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
