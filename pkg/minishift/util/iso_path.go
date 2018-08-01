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
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

func GetIsoPath(isoURL string) string {
	// Make sure isoURL validation happen before using this function.
	// This function ignore any error from url.Parse
	uri, _ := url.Parse(isoURL)

	switch true {
	case strings.Contains(filepath.Base(isoURL), minishiftConstants.B2dIsoAlias):
		return filepath.Join(minishiftConstants.B2dIsoAlias, getIsoVersion(uri.Path))
	case strings.Contains(filepath.Base(isoURL), minishiftConstants.CentOsIsoAlias):
		return filepath.Join(minishiftConstants.CentOsIsoAlias, getIsoVersion(uri.Path))
	default:
		// This handle any random URI
		return filepath.Join("unnamed")
	}

}

func getIsoVersion(isoName string) string {
	re := regexp.MustCompile("v[0-9]*.[0-9]*.[0-9]*")
	return re.FindString(isoName)
}
