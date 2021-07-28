package fs

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func cleanPath(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Fatal(err)
	}
}

func cleanFile(path string) {
	err := os.Remove(path)
	if err != nil {
		log.Fatal(err)
	}
}

func TestCreateMiddlewarePathWithPath(t *testing.T) {
	assert := assert.New(t)
	path := "test_folder"
	expected := fmt.Sprintf("%s%chttp%cmiddleware", path, os.PathSeparator, os.PathSeparator)

	assert.Equal(expected, CreateMiddlewarePath(path))
}

func TestCreateMiddlewarePathWithoutPath(t *testing.T) {
	assert := assert.New(t)
	expected := fmt.Sprintf("http%cmiddleware", os.PathSeparator)

	assert.Equal(expected, CreateMiddlewarePath(""))

}

func TestCreateResourceFileWithPath(t *testing.T) {
	assert := assert.New(t)
	path := "test_folder"
	content := "this a sample file\nIt's use for test"
	filePath := fmt.Sprintf("%s%c%s", path, os.PathSeparator, "sample.go")

	if err := CreateResourceFile(path, "sample", []byte(content)); err != nil {
		log.Fatal(err)
	}

	resultContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	assert.FileExists(filePath)
	assert.Equal(content, string(resultContent))

	cleanPath(path)
}

func TestCreateResourceFileWithoutPath(t *testing.T) {
	assert := assert.New(t)
	content := "this a sample file\nIt's use for test"

	if err := CreateResourceFile("", "sample", []byte(content)); err != nil {
		log.Fatal(err)
	}

	resultContent, err := os.ReadFile("sample.go")
	if err != nil {
		log.Fatal(err)
	}

	assert.FileExists("sample.go")
	assert.Equal(content, string(resultContent))

	cleanFile("sample.go")
}

func TestCreateResourceFileAlreadyExist(t *testing.T) {
	assert := assert.New(t)
	content := "this a sample file\nIt's use for test"
	errorContent := "File already exists"

	if err := CreateResourceFile("", "sample", []byte(content)); err != nil {
		log.Fatal(err)
	}

	err := CreateResourceFile("", "sample", []byte(content))

	assert.NotNil(err)
	assert.Equal(errorContent, fmt.Sprintf("%s", err))

	cleanFile("sample.go")
}
