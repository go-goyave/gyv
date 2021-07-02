package mod

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ProjectNameFromModuleName is a function which extract project name from a module name
func ProjectNameFromModuleName(moduleName *string) string {
	return strings.Split(*moduleName, "/")[bytes.Count([]byte(*moduleName), []byte("/"))]
}

// ReplaceAll is a function which replace all default module names or package names by injected values
func ReplaceAll(projectName string, moduleName string) error {
	if err := replaceProjectModuleName(projectName, moduleName); err != nil {
		return err
	}
	if err := replaceGoModPackageName(projectName, moduleName); err != nil {
		return err
	}

	return nil
}

func replaceProjectModuleName(projectName string, moduleName string) error {
	return filepath.Walk(projectName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		result := string(bytes)
		for _, defaulValue := range defaultTemplateValues() {
			result = strings.ReplaceAll(result, defaulValue, moduleName)

		}

		if err := ioutil.WriteFile(path, []byte(result), 0644); err != nil {
			return err
		}

		return nil
	})
}

func defaultTemplateValues() []string {
	return []string{"goyave.dev/template", "goyave_template"}
}

func replaceGoModPackageName(projectName string, moduleName string) error {
	goModPath := fmt.Sprintf(".%c%s%cgo.mod", os.PathSeparator, projectName, os.PathSeparator)
	goModBytes, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return err
	}

	result := string(goModBytes)
	for _, defaultValue := range defaultTemplateValues() {
		result = strings.ReplaceAll(result, fmt.Sprintf("module %s", defaultValue), fmt.Sprintf("module %s", moduleName))
	}

	if err := ioutil.WriteFile(goModPath, []byte(result), 0644); err != nil {
		return err
	}

	return nil
}
