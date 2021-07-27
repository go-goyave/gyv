package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/Masterminds/semver"
)

// CreateResourceFile create a resource file from a stub
func CreateResourceFile(path string, name string, data []byte) error {
	var filePath string
	if path == "" {
		filePath = fmt.Sprintf("%s.go", name)
	} else {
		filePath = fmt.Sprintf("%s%c%s.go", path, os.PathSeparator, name)
		if err := CreatePath(path); err != nil {
			return err
		}
	}

	info, err := os.Stat(filePath)
	if info != nil {
		return fmt.Errorf("File already exists")
	}
	if !os.IsNotExist(err) {
		return err
	}

	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// CreateControllerPath generate the path to goyave controllers according to the version
func CreateControllerPath(controllerName string, projectPath string) (string, error) {
	goyaveVersion, err := GetGoyaveVersion(projectPath)
	if err != nil {
		return "", err
	}

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

// CreatePath create a directory and its parents if necessary
func CreatePath(path string) error {
	info, err := os.Stat(path)
	if info != nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(path, 0744); err != nil {
		return err
	}

	return nil
}

// CreateMiddlewarePath generate the path to Goyave middleware according to the version
func CreateMiddlewarePath(projectPath string) string {
	if projectPath == "" {
		return fmt.Sprintf("http%cmiddleware", os.PathSeparator)
	}

	return fmt.Sprintf("%s%chttp%cmiddleware", projectPath, os.PathSeparator, os.PathSeparator)
}

// CreateModelPath generate the path to Goyave models according to the version
func CreateModelPath(modelName string, projectPath string) (string, error) {
	goyaveVersion, err := GetGoyaveVersion(projectPath)
	if err != nil {
		return "", err
	}

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

// CopyFile copy file from given source path to given destination path
func CopyFile(source, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}

	out, err := os.Create(destination)
	if err != nil {
		_ = in.Close()
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		_ = in.Close()
		_ = out.Close()
		return err
	}

	_ = in.Close()
	return out.Close()
}
