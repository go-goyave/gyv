package fs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	createPathSample = "parent/child"
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

func TestCreatePath(t *testing.T) {
	assert := assert.New(t)

	if err := CreatePath(createPathSample); err != nil {
		log.Fatal(err)
	}

	assert.DirExists(createPathSample)
	cleanPath("parent")
}

func TestCreatePathAlreadyExist(t *testing.T) {
	assert := assert.New(t)
	sampleFile := fmt.Sprintf("%s%c%s", createPathSample, os.PathSeparator, "file.go")

	if err := CreatePath(createPathSample); err != nil {
		assert.FailNow(err.Error())
	}

	file, err := os.Create(sampleFile)
	if err != nil {
		assert.FailNow(err.Error())
	}
	assert.Nil(file.Close())

	if err := CreatePath(createPathSample); err != nil {
		assert.FailNow(err.Error())
	}

	assert.DirExists(createPathSample)
	assert.FileExists(sampleFile)

	cleanPath("parent")
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

	resultContent, err := ioutil.ReadFile(filePath)
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

	resultContent, err := ioutil.ReadFile("sample.go")
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
	errorContent := "File already exist"

	if err := CreateResourceFile("", "sample", []byte(content)); err != nil {
		log.Fatal(err)
	}

	err := CreateResourceFile("", "sample", []byte(content))

	assert.NotNil(err)
	assert.Equal(errorContent, fmt.Sprintf("%s", err))

	cleanFile("sample.go")
}
