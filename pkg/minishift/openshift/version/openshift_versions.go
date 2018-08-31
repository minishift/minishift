/*
Copyright (C) 2016 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package version

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/minishift/minishift/pkg/util/github"
	"github.com/minishift/minishift/pkg/version"
)

func GetOpenshiftVersion(sshCommander provision.SSHCommander) (string, error) {
	dockerCommander := docker.NewVmDockerCommander(sshCommander)
	return dockerCommander.Exec(" ", minishiftConstants.OpenshiftContainerName, "openshift", "version")
}

// PrintUpstreamVersions prints the origin versions which satisfies the following conditions:
// 	1. Major versions greater than or equal to the minimum supported and default version
//	2. Pre-release versions greater than default version
func PrintUpstreamVersions(output io.Writer, minSupportedVersion string, defaultVersion string) error {
	tags, err := GetGithubReleases()
	if err != nil {
		return err
	}

	releaseList, err := OpenShiftTagsByAscending(tags, minSupportedVersion, defaultVersion)
	if err != nil {
		return err
	}

	fmt.Fprint(output, "The following OpenShift versions are available: \n")
	for _, tag := range releaseList {
		fmt.Fprintf(output, "\t- %s\n", tag)
	}
	return nil
}

func getResponseBody(url string) (resp *http.Response, err error) {
	resp, err = http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func isPrerelease(tag string) bool {
	if match, _ := regexp.MatchString("alpha|beta|rc", tag); match {
		return true
	}
	return false
}

// IsGreaterOrEqualToBaseVersion returns true if the version is greater or equal to the base version
func IsGreaterOrEqualToBaseVersion(version string, baseVersion string) (bool, error) {
	v, err := semver.Parse(strings.TrimPrefix(version, constants.VersionPrefix))
	if err != nil {
		return false, errors.New(fmt.Sprintf("Invalid version format '%s': %s", version, err.Error()))
	}

	baseVersionToCompare := strings.TrimPrefix(baseVersion, constants.VersionPrefix)
	versionRange, err := semver.ParseRange(fmt.Sprintf(">=%s", baseVersionToCompare))
	if err != nil {
		fmt.Println("Not able to parse version info", err)
		return false, err
	}

	if versionRange(v) {
		return true, nil
	}
	return false, nil
}

func GetGithubReleases() ([]string, error) {
	var releaseTags []string
	ctx := context.Background()
	client := github.Client()
	listOptions := github.ListOptions()
	releases, _, err := client.Repositories.ListReleases(ctx, "openshift", "origin", listOptions)
	if err != nil {
		if github.IsRateLimitError(err) {
			return nil, fmt.Errorf("Hit github rate limit: %v", err)
		}
		return nil, err
	}
	for _, release := range releases {
		releaseTags = append(releaseTags, *release.Name)
	}
	return releaseTags, nil
}

func OpenShiftTagsByAscending(tags []string, minSupportedVersion, defaultVersion string) ([]string, error) {
	var tagList []string

	for _, tag := range tags {
		if valid, _ := IsGreaterOrEqualToBaseVersion(tag, minSupportedVersion); valid {
			if valid, _ := IsGreaterOrEqualToBaseVersion(tag, defaultVersion); valid {
				tagList = append(tagList, tag)
			} else {
				if !isPrerelease(tag) {
					tagList = append(tagList, tag)
				}
			}
		}
	}

	return sortTagsViaSemverSort(tagList), nil
}

func sortTagsViaSemverSort(tags []string) []string {
	var (
		versionTags   []semver.Version
		tagsInStrings []string
	)

	for _, tag := range tags {
		semVerTag, _ := semver.Parse(strings.TrimPrefix(tag, version.VersionPrefix))
		versionTags = append(versionTags, semVerTag)
	}

	semver.Sort(versionTags)

	// again apply prefix to tag
	for _, tag := range versionTags {
		tagsInStrings = append(tagsInStrings, fmt.Sprintf("v%s", tag))
	}

	return tagsInStrings
}
