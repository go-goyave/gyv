package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ExtractZip unzip a zip file
func ExtractZip(filename string, projectName string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(filename)
	if err != nil {
		return filenames, err
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Println(err)
		}
	}()

	for _, f := range r.File[1:] {
		fpath := filepath.Join(projectName, strings.Join(strings.Split(f.Name, "/")[1:], "/"))

		if !strings.HasPrefix(fpath, filepath.Clean(projectName)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return nil, err
			}
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return filenames, err
		}

		err = outFile.Close()
		if err != nil {
			return filenames, err
		}

		err = rc.Close()
		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
