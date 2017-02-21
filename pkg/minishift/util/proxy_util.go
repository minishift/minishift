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
	"regexp"
	"strings"

	"github.com/golang/glog"
)

var (
	proxyUserInfoSlice []string
	proxyUriSlice      []string
	matched            bool
	err                error
)

// Parses the proxy URI and will return  ["proxy_hostname", "proxy_port", "user", "password"]
// For unauthenticated proxy server it will return ["proxy_hostname:proxy_port", "user"]
// Or ["proxy_hostname:proxy_port"] depending upon the URI
func ParseProxyUri(proxyUri string) ([]string, error) {
	if matched, err = regexp.MatchString("http://", proxyUri); matched {
		proxyUri = strings.TrimPrefix(proxyUri, "http://")
	} else if matched, err = regexp.MatchString("https://", proxyUri); matched {
		proxyUri = strings.TrimPrefix(proxyUri, "https://")
	}

	if err != nil {
		glog.Errorf("Error starting the VM: %s. Retrying.\n", err)
		return nil, err
	}

	// Get rid of the trailing "/" if any
	proxyUri = strings.TrimSuffix(proxyUri, "/")

	// Value of proxyUri at this time would be similar to fedora:fedora123@xyz.com:8000
	// splitting the string at the last occurence of "@". As we know proxy URL will not contain "@"
	index := strings.LastIndex(proxyUri, "@")

	// For an unauthenticated proxy username and password will not be present
	if index != -1 {
		proxyUserInfo := strings.TrimSuffix(proxyUri[0:index+1], "@")
		proxyUserInfoSlice = strings.Split(proxyUserInfo, ":")
	}

	proxyUrl := proxyUri[index+1:]
	proxyUrlSlice := strings.Split(proxyUrl, ":")

	if proxyUserInfoSlice != nil {
		proxyUriSlice = append(proxyUrlSlice, proxyUserInfoSlice...)
		return proxyUriSlice, nil
	}
	return proxyUrlSlice, nil
}
