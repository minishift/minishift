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
	"fmt"
	"github.com/minishift/minishift/pkg/util"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

type ImageTags struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []Results   `json:"results"`
}

type Results struct {
	Name        string      `json:"name"`
	FullSize    int         `json:"full_size"`
	Images      []ImageInfo `json:"images"`
	ID          int         `json:"id"`
	Repository  int         `json:"repository"`
	Creator     int         `json:"creator"`
	LastUpdater int         `json:"last_updater"`
	LastUpdated time.Time   `json:"last_updated"`
	ImageID     interface{} `json:"image_id"`
	V2          bool        `json:"v2"`
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

func PrintDownStreamVersions(output io.Writer, minSupportedVersion string) {
	resp, err := getResponseBody("https://registry.access.redhat.com/v1/repositories/openshift3/ose/tags")
	if err != nil {
		fmt.Println("Error Occured", err)
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var data map[string]string
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Printf("%T\n%s\n%#v\n", err, err, err)
		return
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
}

func PrintUpStreamVersions(output io.Writer, minSupportedVersion string) {
	resp, err := getResponseBody("https://registry.hub.docker.com/v2/repositories/openshift/origin/tags/")
	if err != nil {
		fmt.Fprintf(output, "Error Occured %s", err)
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var data ImageTags
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Printf("%T\n%s\n%#v\n", err, err, err)
		return
	}
	fmt.Fprint(output, "The following OpenShift versions are available: \n")
	var tagsList []string
	for _, imageinfo := range data.Results {
		if strings.Contains(imageinfo.Name, "latest") {
			continue
		}
		if util.VersionOrdinal(imageinfo.Name) >= util.VersionOrdinal(minSupportedVersion) {
			tagsList = append(tagsList, imageinfo.Name)
		}
	}
	sort.Strings(tagsList)
	for _, tag := range tagsList {
		fmt.Fprintf(output, "\t- %s\n", tag)
	}
}

func getResponseBody(url string) (resp *http.Response, err error) {
	resp, err = http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
