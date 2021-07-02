package create

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"goyave.dev/gyv/command"
	"goyave.dev/gyv/fs"
	"goyave.dev/gyv/stub"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	defaultModName = "models"
)

type CreateModel struct {
	ModelName   string
	ProjectPath string
}

func (c *CreateModel) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Create goyave model",
		Long: `Command for create goyave model with the help of a survey.
Only the model-name flag is required. The project-path is optional.
Is project-path is empty, the program going to search for a goyave project root.
And for this, the program going to move up to the parent directory and check each time if this directory is a goyave project.
If any parents directories are goyave project, an error will be raised.`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

func (c *CreateModel) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{
		{
			Name:     "modelName",
			Prompt:   &survey.Input{Message: "Input the name of the model"},
			Validate: survey.Required,
		},
		{
			Name:   "projectPath",
			Prompt: &survey.Input{Message: "Input the path to the goyave project"},
		},
	}, nil
}

func (c *CreateModel) Execute() error {
	if err := fs.IsValidProject(c.ProjectPath); err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	goyaveModVersion, err := fs.GetGoyavePath(c.ProjectPath)
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	goyaveVersion, err := fs.GetGoyaveVersion(c.ProjectPath)
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	stubPath, err := stub.GenerateStubVersionPath(stub.Model, *goyaveVersion)
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	templateData, err := stub.Load(*stubPath, stub.Data{
		"GoyaveModVersion": *goyaveModVersion,
		"ModelName":        strings.Title(c.ModelName),
	})
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	folderPath, err := fs.CreateModelPath(c.ModelName, c.ProjectPath)
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	err = fs.CreateResourceFile(*folderPath, c.ModelName, templateData.Bytes())
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	fmt.Println("✅ File Created !")

	return nil
}

func (c *CreateModel) Validate() error {
	if c.ModelName == "" {
		return errors.New("❌ required flag \"model-name\"")
	}

	return nil
}

func (c *CreateModel) UsedFlags() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--model-name" || arg == "-n" {
			return true
		}
	}

	return false
}

func (c *CreateModel) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.ModelName,
		"model-name",
		"n",
		"",
		"The name of the model to generate",
	)
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the goyave project root",
	)
}
