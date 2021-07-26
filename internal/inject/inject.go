package inject

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// read main.go and find the main route registrer using AST
// can be another file that is not named main.go (find the main function and where goyave.Start is located)

// Create a temp file with current timestamp in name
// containing a function that returns the main route registrer
// Compile in plugin mode
// Use the injected function to generate a router
// Delete the tmp file
// Delete the built plugin (or output it to /tmp?)

func generateTempFileName() string {
	return fmt.Sprintf("codeinject-%d.go", time.Now().Unix())
}

func generatePluginPath() string {
	return fmt.Sprintf("%s%cgyv-code-injection.go", os.TempDir(), os.PathSeparator)
}

func buildPlugin(output string) error {
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", output)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
