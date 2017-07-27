/*
Copyright (C) 2017 Red Hat, Inc.

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

package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type VersionSpec struct {
	Assets []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func GetLatestISOVersion(isoName string) string {
	resp, err := getResponseBody(fmt.Sprintf("https://api.github.com/repos/minishift/minishift-%s-iso/releases/latest", isoName))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var versionSpec VersionSpec
	err = decoder.Decode(&versionSpec)
	if err != nil {
		return ""
	}
	for _, val := range versionSpec.Assets {
		if strings.HasSuffix(val.Name, ".iso") {
			return val.BrowserDownloadURL
		}
	}
	return ""
}

func getResponseBody(url string) (resp *http.Response, err error) {
	resp, err = http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
