package create

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"goyave.dev/gyv/internal/command"
	"goyave.dev/gyv/internal/fs"
	"goyave.dev/gyv/internal/stub"
)

// ModelData the data injected by the user to generate a model
type ModelData struct {
	ModelName   string
	ProjectPath string
}

// BuildCobraCommand builds the cobra command for this action
func (c *ModelData) BuildCobraCommand() *cobra.Command {
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
func (c *ModelData) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{
		{
			Name:     "modelName",
			Prompt:   &survey.Input{Message: "Model name"},
			Validate: survey.Required,
		},
		{
			Name:   "projectPath",
			Prompt: &survey.Input{Message: "Project path (leave empty for auto-detect)"},
		},
	}, nil
}

// Execute the command's behavior
func (c *ModelData) Execute() error {
	if err := fs.IsValidProject(c.ProjectPath); err != nil {
		return err
	}

	goyaveModVersion, err := fs.GetGoyavePath(c.ProjectPath)
	if err != nil {
		return err
	}

	goyaveVersion, err := fs.GetGoyaveVersion(c.ProjectPath)
	if err != nil {
		return err
	}

	stubPath, err := stub.GenerateStubVersionPath(stub.Model, goyaveVersion)
	if err != nil {
		return err
	}

	templateData, err := stub.Load(stubPath, stub.Data{
		"GoyaveModVersion": goyaveModVersion,
		"ModelName":        strings.Title(c.ModelName),
	})
	if err != nil {
		return err
	}

	folderPath, err := fs.CreateModelPath(c.ModelName, c.ProjectPath)
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
func (c *ModelData) Validate() error {
	if c.ModelName == "" {
		return errors.New("❌ required flag \"name\"")
	}

	return nil
}

// UsedFlags checks if flags are used
func (c *ModelData) UsedFlags() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--name" || arg == "-n" {
			return true
		}
	}

	return false
}

func (c *ModelData) setFlags(flags *pflag.FlagSet) {
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
