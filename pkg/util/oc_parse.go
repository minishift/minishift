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

// Use to parse oc cluster up -h options.

package util

import (
	"regexp"
	"strings"
)

func ParseOcHelpCommand(cmdOut []byte) []string {
	ocOptions := []string{}
	ocOptionRegex := regexp.MustCompile(`(?s)Options(.*)OpenShift images`)
	matches := ocOptionRegex.FindSubmatch(cmdOut)
	if matches != nil {
		tmpOptionsList := string(matches[0])
		for _, value := range strings.Split(tmpOptionsList, "\n")[1:] {
			tmpOption := strings.Split(strings.Split(strings.TrimSpace(value), "=")[0], "--")
			if len(tmpOption) > 1 {
				ocOptions = append(ocOptions, tmpOption[1])
			}
		}
	} else {
		return nil
	}
	return ocOptions
}

func FlagExist(ocCommandOptions []string, flag string) bool {
	for _, v := range ocCommandOptions {
		if v == flag {
			return true
		}
	}
	return false
}
