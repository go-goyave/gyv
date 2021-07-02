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

type CreateProject struct {
	GoyaveVersion string
	ModuleName    string
}

func (c *CreateProject) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Create goyave project",
		Long: `Command for create goyave project with the help of a survey.
Some go and git commands are used for create the project, go and git must be installed and definded in your environment.
The flags --module-name and --goyave-version are required.
In the case where you don't use flags, a survey will be start to allow you to inject the data.`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

func (c *CreateProject) BuildSurvey() ([]*survey.Question, error) {
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

func getTagsName(tags []git.GitTag) []string {
	var tagsName []string

	for _, tag := range tags {
		tagsName = append(tagsName, tag.Name)
	}

	return tagsName
}

func (c *CreateProject) Execute() error {
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

	if err := git.GitInit(projectName); err != nil {
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

func (c *CreateProject) Validate() error {
	if c.GoyaveVersion == "" || c.ModuleName == "" {
		return errors.New("required flag(s) \"goyave-version\" and \"module-name\" aren't set")
	}

	return nil
}

func (c *CreateProject) UsedFlags() bool {
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

func (c *CreateProject) setFlags(flags *pflag.FlagSet) {
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
