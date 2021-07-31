package create

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"goyave.dev/gyv/internal/command"
	"goyave.dev/gyv/internal/fs"
	"goyave.dev/gyv/internal/stub"
)

// Model command for model generation
type Model struct {
	command.ProjectPathCommand
	ModelName string
}

// BuildCobraCommand builds the cobra command for this action
func (c *Model) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Create a Goyave model",
		Long: `Command to create Goyave model.
Only the model-name flag is required. The project-path is optional.
If project-path is not specified, the nearest directory containing a go.mod file importing Goyave will be used.`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

// BuildSurvey builds a survey for this action
func (c *Model) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{
		{
			Name:     "ModelName",
			Prompt:   &survey.Input{Message: "Model name"},
			Validate: survey.Required,
		},
	}, nil
}

// Execute the command's behavior
func (c *Model) Execute() error {
	stubPath, err := stub.GenerateStubVersionPath(stub.Model, c.GoyaveVersion)
	if err != nil {
		return err
	}

	templateData, err := stub.Load(stubPath, stub.Data{
		"GoyaveImportPath": c.GoyaveMod.Mod.Path,
		"ModelName":        strings.Title(c.ModelName),
	})
	if err != nil {
		return err
	}

	folderPath, err := fs.CreateModelPath(c.ModelName, c.ProjectPath, c.GoyaveVersion)
	if err != nil {
		return err
	}

	err = fs.CreateResourceFile(folderPath, c.ModelName, templateData.Bytes())
	if err != nil {
		return err
	}

	fmt.Println("✅ Model created!")

	return nil
}

// Validate checks if required flags are definded
func (c *Model) Validate() error {
	if c.ModelName == "" {
		return errors.New("❌ required flag \"name\"")
	}

	return nil
}

func (c *Model) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.ModelName,
		"name",
		"n",
		"",
		"The name of the model to generate",
	)
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the Goyave project root",
	)
}
