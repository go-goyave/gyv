package fs

import (
	"fmt"
	"os"

	"github.com/Masterminds/semver"
)

// CreateControllerPath generate the path to goyave controllers according to the version
func CreateControllerPath(controllerName string, projectPath string, goyaveVersion *semver.Version) (string, error) {
	upperThan1, err := semver.NewConstraint("> 1.X.X")
	if err != nil {
		return "", err
	}

	var basePath string
	if projectPath == "" {
		basePath = "http"
	} else {
		basePath = fmt.Sprintf("%s%chttp", projectPath, os.PathSeparator)
	}

	controllerPath := fmt.Sprintf("%s%ccontrollers%c%s", basePath, os.PathSeparator, os.PathSeparator, controllerName)
	if upperThan1.Check(goyaveVersion) {
		controllerPath = fmt.Sprintf("%s%ccontroller%c%s", basePath, os.PathSeparator, os.PathSeparator, controllerName)
		return controllerPath, nil
	}

	return controllerPath, nil
}

// CreateMiddlewarePath generate the path to Goyave middleware according to the version
func CreateMiddlewarePath(projectPath string) string {
	if projectPath == "" {
		return fmt.Sprintf("http%cmiddleware", os.PathSeparator)
	}

	return fmt.Sprintf("%s%chttp%cmiddleware", projectPath, os.PathSeparator, os.PathSeparator)
}

// CreateModelPath generate the path to Goyave models according to the version
func CreateModelPath(modelName string, projectPath string, goyaveVersion *semver.Version) (string, error) {
	upperThan1, err := semver.NewConstraint("> 1.X.X")
	if err != nil {
		return "", err
	}

	var basePath string
	if projectPath == "" {
		basePath = "database"
	} else {
		basePath = fmt.Sprintf("%s%cdatabase", projectPath, os.PathSeparator)
	}

	path := fmt.Sprintf("%s%cmodels", basePath, os.PathSeparator)
	if upperThan1.Check(goyaveVersion) {
		path = fmt.Sprintf("%s%cmodel", basePath, os.PathSeparator)
		return path, nil
	}

	return path, nil
}
