package create

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"goyave.dev/gyv/internal/command"
	"goyave.dev/gyv/internal/fs"
	"goyave.dev/gyv/internal/git"
	"goyave.dev/gyv/internal/mod"
)

const (
	defaultZipFileName = "goyave_template.zip"
)

// ProjectData the data injected by the user to generate a goyave project
type ProjectData struct {
	GoyaveVersion string
	ModuleName    string
}

// BuildCobraCommand builds the cobra command for this action
func (c *ProjectData) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Create a Goyave project",
		Long: `Command to create Goyave project.
You need go and git to be installed on your system in order de run this command.
The flags --module-name and --goyave-version are required.`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

// BuildSurvey builds a survey for this action
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
			Prompt:   &survey.Input{Message: "Go module name"},
			Validate: survey.Required,
		},
		{
			Name: "goyaveVersion",
			Prompt: &survey.Select{
				Message: "Goyave version number",
				Options: git.VersionsToStrings(filteredVersions),
				Default: filteredVersions[0].String(),
			},
		},
	}, nil

}

// Execute the command's behavior
func (c *ProjectData) Execute() error {
	tags, err := git.GetAllTags()
	if err != nil {
		return err
	}

	projectName := mod.ProjectNameFromModuleName(c.ModuleName)

	info, err := os.Stat(projectName)
	if info != nil {
		return errors.New("A directory with this name already exists")
	}
	if !os.IsNotExist(err) {
		return err
	}

	tag, err := git.GetTagByName(c.GoyaveVersion, tags)
	if err != nil {
		return err
	}

	if err := git.DownloadFile(tag.ZipballURL, defaultZipFileName); err != nil {
		return err
	}

	if _, err := fs.ExtractZip(defaultZipFileName, projectName); err != nil {
		return err
	}

	if err := fs.ReplaceAll(projectName, c.ModuleName); err != nil {
		return err
	}

	if err := git.ProjectGitInit(projectName); err != nil {
		return err
	}

	if err := os.Remove(defaultZipFileName); err != nil {
		return err
	}

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}

	tidyCommand := exec.Command("go", "mod", "tidy")
	projectPath := fmt.Sprintf("%s%c%s", currentWorkingDirectory, os.PathSeparator, projectName)
	tidyCommand.Dir = projectPath
	if err := tidyCommand.Run(); err != nil {
		return err
	}

	err = fs.CopyFile(
		fmt.Sprintf("%s%c%s", projectPath, os.PathSeparator, "config.example.json"),
		fmt.Sprintf("%s%c%s", projectPath, os.PathSeparator, "config.json"),
	)
	if err != nil {
		return err
	}

	fmt.Println("✅ Project created!")
	fmt.Printf("➡️ Get started by navigating to \"%s\"\n", projectName)

	return nil
}

// Validate check if required flags are definded
func (c *ProjectData) Validate() error {
	if c.GoyaveVersion == "" || c.ModuleName == "" {
		return errors.New("required flag(s) \"goyave-version\" and \"module-name\" aren't set")
	}

	return nil
}

// UsedFlags check if flags are used
func (c *ProjectData) UsedFlags() bool { // TODO this is redundant and could just use cobra flags
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
		"The Goyave version used in this project",
	)

}
