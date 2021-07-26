package create

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"goyave.dev/gyv/internal/command"
	"goyave.dev/gyv/internal/fs"
	"goyave.dev/gyv/internal/stub"
)

// ControllerData the data injected by the user to generate a controller
type ControllerData struct {
	ControllerName string
	ProjectPath    string
}

// BuildCobraCommand builds the cobra command for this action
func (c *ControllerData) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controller",
		Short: "Create a Goyave controller",
		Long: `Command to create Goyave controller.
Only the controller-name flag is required. The project-path is optional.
If project-path is not specified, the nearest directory containing a go.mod file importing Goyave will be used.
`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

// BuildSurvey builds a survey for this action
func (c *ControllerData) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{
		{
			Name:     "controllerName",
			Prompt:   &survey.Input{Message: "Controller name"},
			Validate: survey.Required,
		},
		{
			Name:   "projectPath",
			Prompt: &survey.Input{Message: "Project path (leave empty for auto-detect)"},
		},
	}, nil
}

// Execute the command's behavior
func (c *ControllerData) Execute() error {
	if err := fs.IsValidProject(c.ProjectPath); err != nil {
		return err
	}

	// TODO extract actual behavior (excluding validation and visual output)
	// That would help "front-end" part of the CLI to be swapped with ease.
	folderPath, err := fs.CreateControllerPath(c.ControllerName, c.ProjectPath)
	if err != nil {
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

	stubPath, err := stub.GenerateStubVersionPath(stub.Controller, *goyaveVersion)
	if err != nil {
		return err
	}

	templateData, err := stub.Load(*stubPath, stub.Data{
		"GoyaveModVersion": goyaveModVersion,
		"ControllerName":   c.ControllerName,
	})
	if err != nil {
		return err
	}

	if err := fs.CreatePath(*folderPath); err != nil {
		return err
	}

	err = fs.CreateResourceFile(*folderPath, c.ControllerName, templateData.Bytes())
	if err != nil {
		return err
	}

	fmt.Println("✅ Controller created!")

	return nil
}

// Validate checks if required flags are definded
func (c *ControllerData) Validate() error {
	if c.ControllerName == "" {
		return errors.New("required flag(s) \"name\"")
	}

	return nil
}

// UsedFlags checks if flags are used
func (c *ControllerData) UsedFlags() bool {
	controllerNameCheck := false

	for _, arg := range os.Args[1:] {
		if arg == "--name" || arg == "-n" {
			controllerNameCheck = true
		}
	}

	return controllerNameCheck
}

func (c *ControllerData) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.ControllerName,
		"name",
		"n",
		"",
		"The name of the controller to generate",
	)
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the Goyave project root",
	)

}
