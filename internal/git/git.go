package git

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver"
)

// Tag represent github api response
type Tag struct {
	Name       string `json:"name"`
	ZipballURL string `json:"zipball_url"`
	TarballURL string `json:"tarball_url"`
	Commit     struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	NodeID string `json:"node_id"`
}

func (tag *Tag) String() string {
	jsonValue, err := json.Marshal(tag)
	if err != nil {
		panic(err)
	}

	return string(jsonValue)
}

// GetVersions convert Tag into semver Version
func GetVersions(tags []Tag) ([]*semver.Version, error) {
	var versions []*semver.Version

	for _, tag := range tags {
		version, err := tag.getVersion()
		if err != nil {
			return nil, err
		}

		versions = append(versions, version)
	}

	return versions, nil
}

func (tag *Tag) getVersion() (*semver.Version, error) {
	return semver.NewVersion(tag.Name)
}

// GetTopVersions return all major and minor versions
// Only the highest patch version is taken
func GetTopVersions(versions []*semver.Version) ([]*semver.Version, error) {
	filteredVersions := versions[:1]
	for _, version := range versions {
		exist, err := majorMinorExist(version, filteredVersions)
		if err != nil {
			return nil, err
		}

		if !exist {
			filteredVersions = append(filteredVersions, version)
			continue
		}

		for index, filteredVersion := range filteredVersions {
			filteredVersionTopMinor := filteredVersion.IncMinor()
			isBigger, err := semver.NewConstraint(fmt.Sprintf("> %s, < %s", filteredVersion.String(), filteredVersionTopMinor.String()))
			if err != nil {
				return nil, err
			}

			if isBigger.Check(version) {
				filteredVersions[index] = version
				break
			}
		}
	}

	return filteredVersions, nil
}

func majorMinorExist(version *semver.Version, versions []*semver.Version) (bool, error) {
	constraint, err := semver.NewConstraint(buildMajorMinorVersionString(version))
	if err != nil {
		return false, err
	}

	for _, v := range versions {
		if constraint.Check(v) {
			return true, nil
		}
	}

	return false, nil
}

func buildMajorMinorVersionString(version *semver.Version) string {
	majorMinor := strings.Split(version.String(), ".")[0:2]
	majorMinor = append(majorMinor, "X")
	resultString := strings.Join(majorMinor, ".")

	return resultString
}

// VersionsToStrings convert semver Versions into a array of strings
func VersionsToStrings(versions []*semver.Version) []string {
	var result []string

	for _, version := range versions {
		result = append(result, version.String())
	}

	return result
}

// GetAllTags return all Goyave tags registered inside Github API
func GetAllTags() ([]Tag, error) {
	goyaveTagsURL := "https://api.github.com/repos/go-goyave/template/tags"
	bodyContent, link, err := getHTTPData(goyaveTagsURL)
	if err != nil {
		return nil, err
	}

	var responsesBytes [][]byte

	responsesBytes = append(responsesBytes, bodyContent)
	linkData, err := getLinksData(responsesBytes, link)
	if err != nil {
		return nil, err
	}

	responsesBytes = append(responsesBytes, linkData...)

	var tags []Tag
	for _, responseBytes := range responsesBytes {
		var tagsResponse []Tag
		if err := json.Unmarshal(responseBytes, &tagsResponse); err != nil {
			return nil, err
		}

		tags = append(tags, tagsResponse...)

	}

	return tags, nil
}

// GetTagByName search a tag from a string version and a list of tags
func GetTagByName(name string, tags []Tag) (*Tag, error) {
	toCheck, err := semver.NewVersion(name)
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		version, err := semver.NewVersion(tag.Name)
		if err != nil {
			return nil, err
		}

		constraint, err := semver.NewConstraint(version.String())
		if err != nil {
			return nil, err
		}

		if constraint.Check(toCheck) {
			return &tag, nil
		}
	}

	return nil, fmt.Errorf("No tag found for: %s", name)
}

// ProjectGitInit initialize a git project
func ProjectGitInit(projectName string) error {
	if _, err := exec.Command("git", "init", projectName).Output(); err != nil {
		return err
	}

	return nil
}
