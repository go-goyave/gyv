package create

import (
	"errors"
	"fmt"
	"os"

	"goyave.dev/gyv/command"
	"goyave.dev/gyv/fs"
	"goyave.dev/gyv/stub"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ControllerData is a structure which represents the data injected by the user to generate a controller
type ControllerData struct {
	ControllerName string
	ProjectPath    string
}

// BuildCobraCommand is a function used to build a cobra command
func (c *ControllerData) BuildCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controller",
		Short: "Create a goyave controller",
		Long: `Command to create goyave controller with the help of a survey.
Only the controller-name flag is required. The project-path is optional.
Is project-path is empty, the program going to search for a goyave project root.
And for this, the program going to move up to the parent directory and check each time if this directory is a goyave project.
If any parents directories are goyave project, an error will be raised.`,
		RunE: command.GenerateRunFunc(c),
	}

	c.setFlags(cmd.Flags())

	return cmd
}

// BuildSurvey is a function used to build a survey
func (c *ControllerData) BuildSurvey() ([]*survey.Question, error) {
	return []*survey.Question{
		{
			Name:     "controllerName",
			Prompt:   &survey.Input{Message: "Input the name of the controller to generate"},
			Validate: survey.Required,
		},
		{
			Name:   "projectPath",
			Prompt: &survey.Input{Message: "Input the path to the goyave project"},
		},
	}, nil
}

// Execute is the core function of the command
func (c *ControllerData) Execute() error {
	if err := fs.IsValidProject(c.ProjectPath); err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	folderPath, err := fs.CreateControllerPath(c.ControllerName, c.ProjectPath)
	if err != nil {
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

	stubPath, err := stub.GenerateStubVersionPath(stub.Controller, *goyaveVersion)
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	templateData, err := stub.Load(*stubPath, stub.Data{
		"GoyaveModVersion": *goyaveModVersion,
		"ControllerName":   c.ControllerName,
	})
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	if err := fs.CreatePath(*folderPath); err != nil {
		return err
	}

	err = fs.CreateResourceFile(*folderPath, c.ControllerName, templateData.Bytes())
	if err != nil {
		return fmt.Errorf("❌ %s", err.Error())
	}

	fmt.Println("✅ File Created !")

	return nil
}

// Validate is a function which check if required flags are definded
func (c *ControllerData) Validate() error {
	if c.ControllerName == "" {
		return errors.New("required flag(s) \"controller-name\"")
	}

	return nil
}

// UsedFlags is a function which check if flags are used
func (c *ControllerData) UsedFlags() bool {
	controllerNameCheck := false

	for _, arg := range os.Args[1:] {
		if arg == "--controller-name" || arg == "-n" {
			controllerNameCheck = true
		}
	}

	return controllerNameCheck
}

func (c *ControllerData) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(
		&c.ControllerName,
		"controller-name",
		"n",
		"",
		"The name of the controller to generate",
	)
	flags.StringVarP(
		&c.ProjectPath,
		"project-path",
		"p",
		"",
		"The path to the goyave project root",
	)

}
