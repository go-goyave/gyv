package git

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/tomnomnom/linkheader"
)

// HTTPClient is an abstraction of the default HTTP client
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

var (
	// GitClient is the client config for HTTP request
	GitClient HTTPClient
)

func init() {
	GitClient = &http.Client{Timeout: 30 * time.Second}
}

func getLinksData(bodyList [][]byte, link *string) ([][]byte, error) {
	if link == nil {
		return bodyList, nil
	}

	bytes, link, err := getHTTPData(link)
	if err != nil {
		return nil, err
	}

	return getLinksData(append(bodyList, bytes), link)
}

func getHTTPData(url *string) ([]byte, *string, error) {
	request, err := http.NewRequest("GET", *url, nil)
	if err != nil {
		return nil, nil, err
	}

	response, err := GitClient.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer response.Body.Close()

	if response.StatusCode > 299 {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, nil, err
		}

		return nil, nil, fmt.Errorf(string(data))
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}

	nextPage := getNextPageURL(strings.Join(response.Header.Values("Link"), ""))

	if nextPage != nil {
		return bytes, nextPage, nil
	}

	return bytes, nil, nil

}

func getNextPageURL(rawLinks string) *string {
	links := linkheader.Parse(rawLinks)

	for _, link := range links {
		if link.Rel == "next" {
			return &link.URL
		}
	}

	return nil
}

// DownloadFile is a function which download a file with a URL and new name for the downloaded file
func DownloadFile(url string, filename string) error {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	response, err := GitClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

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

	defer file.Close()

	bar := progressbar.DefaultBytes(
		response.ContentLength,
		"downloading",
	)

	if _, err := io.Copy(io.MultiWriter(file, bar), response.Body); err != nil {
		return err
	}

	return nil
}
