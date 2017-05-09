package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/bitrise-io/go-utils/fileutil"
)

var (
	pluginSrcURL        = ""
	version             = ""
	osxExecutableName   = ""
	linuxExecutableName = ""
)

// PluginDefinition ...
type PluginDefinition struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Executable  struct {
		OSX   string `yaml:"osx"`
		Linux string `yaml:"linux"`
	}
	TriggerEvent string        `yaml:"trigger"`
	Requirements []Requirement `yaml:"requirements"`
}

// Requirement ...
type Requirement struct {
	Tool       string `yaml:"tool"`
	MinVersion string `yaml:"min_version"`
	MaxVersion string `yaml:"max_version"`
}

func fatalf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
	os.Exit(1)
}

func init() {
	flag.StringVar(&pluginSrcURL, "src", "", "plugin git source url")
	flag.StringVar(&version, "version", "", "plugin version")
	flag.StringVar(&osxExecutableName, "osx_bin", "", "osx binary name")
	flag.StringVar(&linuxExecutableName, "linux_bin", "", "linux binary name")
}

func main() {
	flag.Parse()

	fmt.Printf("Inputs:\n")
	fmt.Printf("  src: %s\n", pluginSrcURL)
	fmt.Printf("  version: %s\n", version)
	fmt.Printf("  osx_bin: %s\n", osxExecutableName)
	fmt.Printf("  linux_bin: %s\n", linuxExecutableName)

	if pluginSrcURL == "" {
		fatalf("plugin git source url (src) not provided")
	}
	if version == "" {
		fatalf("plugin version (version) not provided")
	}
	if osxExecutableName == "" {
		fatalf("osx binary name (osx_bin) not provided")
	}
	if linuxExecutableName == "" {
		fatalf("linux binary name (linux_bin) not provided")
	}

	pluginDefinitionPth := "bitrise-plugin.yml"

	bytes, err := fileutil.ReadBytesFromFile(pluginDefinitionPth)
	if err != nil {
		fatalf("failed to read: %s, error: %s", pluginDefinitionPth, err)
	}

	var definition PluginDefinition

	if err := yaml.Unmarshal(bytes, &definition); err != nil {
		fatalf("failed to unmarshal plugin definition, error: %s", err)
	}

	osxExecutableInstallURL := pluginSrcURL + filepath.Join("/releases/download", version, osxExecutableName)
	linuxExecutableInstallURL := pluginSrcURL + filepath.Join("/releases/download", version, linuxExecutableName)

	definition.Executable.OSX = osxExecutableInstallURL
	definition.Executable.Linux = linuxExecutableInstallURL

	bytes, err = yaml.Marshal(definition)
	if err != nil {
		fatalf("failed to marshal plugin definition, error: %s", err)
	}

	fmt.Printf("\nPlugin definition:\n")
	fmt.Printf("%s\n", string(bytes))

	if err := fileutil.WriteBytesToFile(pluginDefinitionPth, bytes); err != nil {
		fatalf("failed to write plugin definition, error: %s", err)
	}
}
