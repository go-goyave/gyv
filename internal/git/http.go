package git

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/tomnomnom/linkheader"
)

// HTTPClient is an abstraction of the default HTTP client.
// Mainly used to ease testing.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

var (
	// GitClient is the client config for HTTP request
	GitClient HTTPClient = &http.Client{Timeout: 30 * time.Second}
)

func getLinksData(bodyList [][]byte, link string) ([][]byte, error) {
	if link == "" {
		return bodyList, nil
	}

	bytes, link, err := getHTTPData(link)
	if err != nil {
		return nil, err
	}

	return getLinksData(append(bodyList, bytes), link)
}

func getHTTPData(url string) ([]byte, string, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	response, err := GitClient.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	if response.StatusCode > 299 {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, "", err
		}

		return nil, "", fmt.Errorf(string(data))
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}

	return bytes, getNextPageURL(strings.Join(response.Header.Values("Link"), "")), nil

}

func getNextPageURL(rawLinks string) string {
	links := linkheader.Parse(rawLinks)

	for _, link := range links {
		if link.Rel == "next" {
			return link.URL
		}
	}

	return ""
}

// DownloadFile download a file from given URL and writes it to the given filename
func DownloadFile(url string, filename string) error {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	response, err := GitClient.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	if response.StatusCode > 299 {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf(string(data))
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Println(err)
		}
	}()

	bar := progressbar.DefaultBytes(
		response.ContentLength,
		"Downloading",
	)

	if _, err := io.Copy(io.MultiWriter(file, bar), response.Body); err != nil {
		return err
	}

	return nil
}
