package fs

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	templatePackageNames = []string{"goyave.dev/template", "goyave_template"}
)

// ReplaceAll replace all default module names or package names with injected values
func ReplaceAll(projectName string, moduleName string) error {
	return filepath.WalkDir(projectName, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".mod" && ext != ".json" {
			return nil
		}

		bytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		contents := string(bytes)
		shouldBeProcessed := false // Avoid unnecessary IO
		for _, name := range templatePackageNames {
			if strings.Contains(contents, name) {
				shouldBeProcessed = true
				break
			}
		}
		if !shouldBeProcessed {
			return nil
		}

		replaceValue := moduleName
		if strings.HasPrefix(filepath.Base(path), "config.") && ext == ".json" {
			replaceValue = filepath.Base(moduleName)
		}
		for _, name := range templatePackageNames {
			contents = strings.ReplaceAll(contents, name, replaceValue)
		}

		return os.WriteFile(path, []byte(contents), 0644)
	})
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

// CreateResourceFile create a resource file from a stub
func CreateResourceFile(path string, name string, data []byte) error {
	var filePath string
	if path == "" {
		filePath = fmt.Sprintf("%s.go", name)
	} else {
		filePath = fmt.Sprintf("%s%c%s.go", path, os.PathSeparator, name)
		if err := os.MkdirAll(path, 0744); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}
