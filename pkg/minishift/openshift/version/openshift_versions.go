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

	"github.com/minishift/minishift/pkg/minishift/clusterup"
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
		if valid, _ := clusterup.ValidateOpenshiftMinVersion(imageTag.Name, minSupportedVersion); valid {
			if valid, _ := clusterup.ValidateOpenshiftMinVersion(imageTag.Name, defaultVersion); valid {
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
