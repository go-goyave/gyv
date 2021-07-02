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

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

var (
	GitClient HTTPClient
)

func init() {
	GitClient = &http.Client{Timeout: 30 * time.Second}
}

func GetLinksData(bodyList [][]byte, link *string) ([][]byte, error) {
	if link == nil {
		return bodyList, nil
	}

	bytes, link, err := GetHttpData(link)
	if err != nil {
		return nil, err
	}

	return GetLinksData(append(bodyList, bytes), link)
}

func GetHttpData(url *string) ([]byte, *string, error) {
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

	nextPage := getNextPageUrl(strings.Join(response.Header.Values("Link"), ""))

	if nextPage != nil {
		return bytes, nextPage, nil
	}

	return bytes, nil, nil

}

func getNextPageUrl(rawLinks string) *string {
	links := linkheader.Parse(rawLinks)

	for _, link := range links {
		if link.Rel == "next" {
			return &link.URL
		}
	}

	return nil
}

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
