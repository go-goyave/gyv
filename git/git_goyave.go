package git

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver"
)

type GitTag struct {
	Name       string `json:"name"`
	ZipballURL string `json:"zipball_url"`
	TarballURL string `json:"tarball_url"`
	Commit     struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	NodeID string `json:"node_id"`
}

func (tag *GitTag) String() string {
	jsonValue, err := json.Marshal(tag)
	if err != nil {
		panic(err)
	}

	return string(jsonValue)
}

func GetVersions(tags []GitTag) ([]*semver.Version, error) {
	var versions []*semver.Version

	for _, tag := range tags {
		version, err := tag.GetVersion()
		if err != nil {
			return nil, err
		}

		versions = append(versions, version)
	}

	return versions, nil
}

func (tag *GitTag) GetVersion() (*semver.Version, error) {
	return semver.NewVersion(tag.Name)
}

func GetGitTagByName(tags []GitTag, version *semver.Version) *GitTag {
	for _, tag := range tags {
		if version.String() == tag.Name {
			return &tag
		}
	}

	return nil
}

func GetTopVersions(versions []*semver.Version) ([]*semver.Version, error) {
	filteredVersions := versions[:1]
	for _, version := range versions {
		exist, err := MajorMinorExist(version, filteredVersions)
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

func MajorMinorExist(version *semver.Version, versions []*semver.Version) (bool, error) {
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

func VersionsToStrings(versions []*semver.Version) []string {
	var result []string

	for _, version := range versions {
		result = append(result, version.String())
	}

	return result
}

func GetAllTags() ([]GitTag, error) {
	goyaveTagsURL := "https://api.github.com/repos/go-goyave/template/tags"
	bodyContent, link, err := GetHttpData(&goyaveTagsURL)
	if err != nil {
		return nil, err
	}

	var responsesBytes [][]byte

	responsesBytes = append(responsesBytes, bodyContent)
	linkData, err := GetLinksData(responsesBytes, link)
	if err != nil {
		return nil, err
	}

	responsesBytes = append(responsesBytes, linkData...)

	var tags []GitTag
	for _, responseBytes := range responsesBytes {
		var tagsResponse []GitTag
		if err := json.Unmarshal(responseBytes, &tagsResponse); err != nil {
			return nil, err
		}

		tags = append(tags, tagsResponse...)

	}

	return tags, nil
}

func GetTagByName(name string, tags []GitTag) (*GitTag, error) {
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

func GitInit(projectName string) error {
	if _, err := exec.Command("git", "init", projectName).Output(); err != nil {
		return err
	}

	return nil
}
