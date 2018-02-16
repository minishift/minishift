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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minikubeConstants "github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/minishift/minishift/pkg/util"
)

type ImageTags struct {
	Name string `json:"name"`
}

type ImageInfo struct {
	Size         int         `json:"size"`
	Architecture string      `json:"architecture"`
	Variant      interface{} `json:"variant"`
	Features     interface{} `json:"features"`
	Os           interface{} `json:"os"`
	OsVersion    interface{} `json:"os_version"`
	OsFeatures   interface{} `json:"os_features"`
}

func GetOpenshiftVersion(sshCommander provision.SSHCommander) (string, error) {
	dockerCommander := docker.NewVmDockerCommander(sshCommander)
	return dockerCommander.Exec(" ", minishiftConstants.OpenshiftContainerName, "openshift", "version")
}

func GetOpenshiftVersionWithoutK8sAndEtcd(sshCommander provision.SSHCommander) (string, error) {
	versionInfo, err := GetOpenshiftVersion(sshCommander)
	if err != nil {
		return "", err
	}

	// versionInfo variable have below string as value along with new line
	// openshift v3.6.1+c4dd4cf
	// kubernetes v1.6.1+5115d708d7
	// etcd 3.2.1
	// openShiftVersionAlongWithCommitSha is contain *v3.6.1+c4dd4cf* (first split on new line and second on space)
	openShiftVersionAlongWithCommitSha := strings.Split(strings.Split(versionInfo, "\n")[0], " ")[1]
	// openshiftVersion is contain *3.6.1* (split on *+* string and then trim the *v* as perfix)
	// TrimSpace is there to make sure no whitespace around version string
	openShiftVersion := strings.TrimSpace(strings.TrimPrefix(strings.Split(openShiftVersionAlongWithCommitSha, "+")[0], minikubeConstants.VersionPrefix))
	return openShiftVersion, nil
}

func PrintDownStreamVersions(output io.Writer, minSupportedVersion string) error {
	resp, err := getResponseBody("https://registry.access.redhat.com/v1/repositories/openshift3/ose/tags")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var data map[string]string
	err = decoder.Decode(&data)
	if err != nil {
		return errors.New(fmt.Sprintf("%T\n%s\n%#v\n", err, err, err))
	}
	fmt.Fprint(output, "The following OpenShift versions are available: \n")
	var tagsList []string
	for version := range data {
		if util.VersionOrdinal(version) >= util.VersionOrdinal(minSupportedVersion) {
			if strings.Contains(version, "latest") {
				continue
			}
			if strings.Contains(version, "-") {
				continue
			}
			tagsList = append(tagsList, version)
		}
	}
	sort.Strings(tagsList)
	for _, tag := range tagsList {
		fmt.Fprintf(output, "\t- %s\n", tag)
	}
	return nil
}

func PrintUpStreamVersions(output io.Writer, minSupportedVersion string, defaultVersion string) error {
	dockerRegistryUrl := "https://registry.hub.docker.com/v1/repositories/openshift/origin/tags"
	resp, err := getResponseBody(dockerRegistryUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var data []ImageTags
	err = decoder.Decode(&data)
	if err != nil {
		return errors.New(fmt.Sprintf("%T\n%s\n%#v\n", err, err, err))
	}
	fmt.Fprint(output, "The following OpenShift versions are available: \n")
	var tagsList []string
	for _, imageTag := range data {
		if strings.Contains(imageTag.Name, "latest") {
			continue
		}
		if valid, _ := IsGreaterOrEqualToBaseVersion(imageTag.Name, minSupportedVersion); valid {
			if valid, _ := IsGreaterOrEqualToBaseVersion(imageTag.Name, defaultVersion); valid {
				tagsList = append(tagsList, imageTag.Name)
			} else {
				if !isPrerelease(imageTag.Name) {
					tagsList = append(tagsList, imageTag.Name)
				}
			}
		}
	}
	sort.Strings(tagsList)
	for _, tag := range tagsList {
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
