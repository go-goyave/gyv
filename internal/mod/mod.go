package mod

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/mod/modfile"
)

const (
	goModFilename = "go.mod"
)

var (
	// ErrNotAGoyaveProject returned when the go project found doesn't import Goyave
	ErrNotAGoyaveProject = errors.New("Current project doesn't import Goyave")

	// ErrNoGoMod returned when no go.mod file can be found
	ErrNoGoMod = errors.New("No go.mod found")

	goyaveImportPaths = []string{"goyave.dev/goyave", "github.com/System-Glitch/goyave"}
)

// Parse reads "go.mod" file from the given directory if it exists.
func Parse(directory string) (*modfile.File, error) {
	goModPath := goModFilename

	if directory != "" {
		goModPath = fmt.Sprintf("%s%c%s", directory, os.PathSeparator, goModFilename)
	}

	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}

	modfile, err := modfile.Parse(fmt.Sprintf("%s%c%s", goModPath, os.PathSeparator, goModFilename), data, nil)
	if err != nil {
		return nil, err
	}

	return modfile, nil
}

// FindDependency find the required dependency identified by the given dependencyPath
// (like "golang.org/x/text" or "rsc.io/quote/v2") in the given modFile requires, or nil.
func FindDependency(modFile *modfile.File, dependencyPath string) *modfile.Require {
	for _, d := range modFile.Require {
		if d.Mod.Path == dependencyPath {
			return d
		}
	}
	return nil
}

// SetRootWorkingDirectory set the working directory to the nearest
// directory containing a "go.mod" file (ascending in the directory tree)
// and return that path.
// If there is no matching directory, ErrNoGoMod is returned.
func SetRootWorkingDirectory() (string, error) {
	projectRoot := FindParentModule()
	if projectRoot == "" {
		return "", ErrNoGoMod
	}

	return projectRoot, os.Chdir(projectRoot)
}

// FindGoyaveRequire find the first Goyave occurrence in the given
// modFile's requires, or nil.
func FindGoyaveRequire(modFile *modfile.File) *modfile.Require {
	for _, d := range modFile.Require {
		for _, path := range goyaveImportPaths {
			if strings.HasPrefix(d.Mod.Path, path) {
				return d
			}
		}
	}

	return nil
}

// FindParentModule tries to find the ascending relative path
// to the nearest directory containing a "go.mod" file, or an
// empty string.
func FindParentModule() string {
	sep := string(os.PathSeparator)
	directory, err := os.Getwd()
	if err != nil {
		return ""
	}

	for !fileExists(fmt.Sprintf("%s%s%s", directory, sep, goModFilename)) {
		directory = directory[:strings.LastIndex(directory, sep)]
		if !isDirectory(directory) {
			return ""
		}
	}

	return directory
}

func fileExists(name string) bool {
	if stats, err := os.Stat(goModFilename); err == nil {
		return !stats.IsDir()
	}
	return false
}

func isDirectory(path string) bool {
	if stats, err := os.Stat(path); err == nil {
		return stats.IsDir()
	}
	return false
}

// ProjectNameFromModuleName extract project name from a module name
func ProjectNameFromModuleName(moduleName string) string {
	return strings.Split(moduleName, "/")[bytes.Count([]byte(moduleName), []byte("/"))]
}
