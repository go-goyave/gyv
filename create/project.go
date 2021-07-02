package create

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"goyave.dev/gyv/command"
	"goyave.dev/gyv/fs"
	"goyave.dev/gyv/git"
	"goyave.dev/gyv/mod"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	defaultZipFileName = "goyave_template.zip"
)

// ProjectData is a structure which represents the data injected by the user to generate a goyave project
type ProjectData struct {
	GoyaveVersion string
	ModuleName    string
}

// BuildCobraCommand is a function used to build a cobra command
func (c *ProjectData) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Create a goyave project",
		Long: `Command to create goyave project with the help of a survey.
Some go and git commands are used for create the project, go and git must be installed and definded in your environment.
The flags --module-name and --goyave-version are required.
In the case where you don't use flags, a survey will be start to allow you to inject the data.`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

// BuildSurvey is a function used to build a survey
func (c *ProjectData) BuildSurvey() ([]*survey.Question, error) {
	tags, err := git.GetAllTags()
	if err != nil {
		return nil, err
	}

	versions, err := git.GetVersions(tags)
	if err != nil {
		return nil, err
	}

	filteredVersions, err := git.GetTopVersions(versions)
	if err != nil {
		return nil, err
	}

	return []*survey.Question{
		{
			Name:     "moduleName",
			Prompt:   &survey.Input{Message: "Input the name of the go module"},
			Validate: survey.Required,
		},
		{
			Name: "goyaveVersion",
			Prompt: &survey.Select{
				Message: "Choise Goyave version number:",
				Options: git.VersionsToStrings(filteredVersions),
				Default: filteredVersions[0].String(),
			},
		},
	}, nil

}

// Execute is the core function of the command
func (c *ProjectData) Execute() error {
	tags, err := git.GetAllTags()
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	projectName := mod.ProjectNameFromModuleName(&c.ModuleName)

	info, err := os.Stat(projectName)
	if info != nil {
		return errors.New("❌ A directory with this name already exists")
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("❌ %s", err.Error())
	}

	tag, err := git.GetTagByName(c.GoyaveVersion, tags)
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	if err := git.DownloadFile(tag.ZipballURL, defaultZipFileName); err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	if _, err := fs.ExtractZip(defaultZipFileName, projectName); err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	if err := mod.ReplaceAll(projectName, c.ModuleName); err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	if err := git.ProjectGitInit(projectName); err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	if err := os.Remove(defaultZipFileName); err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	tidyCommand := exec.Command("go", "mod", "tidy")
	tidyCommand.Dir = fmt.Sprintf("%s%c%s", currentWorkingDirectory, os.PathSeparator, projectName)
	if err := tidyCommand.Run(); err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	fmt.Println("✅ Project created !")

	return nil
}

// Validate is a function which check if required flags are definded
func (c *ProjectData) Validate() error {
	if c.GoyaveVersion == "" || c.ModuleName == "" {
		return errors.New("required flag(s) \"goyave-version\" and \"module-name\" aren't set")
	}

	return nil
}

// UsedFlags is a function which check if flags are used
func (c *ProjectData) UsedFlags() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--module-name" || arg == "-n" {
			return true
		}

		if arg == "--goyave-version" || arg == "-g" {
			return true
		}
	}

	return false
}

func (c *ProjectData) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.ModuleName,
		"module-name",
		"n",
		"",
		"The name of your module",
	)
	flags.StringVarP(
		&c.GoyaveVersion,
		"goyave-version",
		"g",
		"",
		"The version of goyave for generate a corresponding template",
	)

}
