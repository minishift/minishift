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

package addon

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var requiredMetaTags = []string{NameMetaTagName, DescriptionMetaTagName}

const (
	requiredVars             = "Required-Vars"
	NameMetaTagName          = "Name"
	DescriptionMetaTagName   = "Description"
	RequiredOpenShiftVersion = "OpenShift-Version"
	anyOpenShiftVersion      = ""
)

// AddOnMeta defines a set of meta data for an AddOn. Name and Description are required. Others are optional.
type AddOnMeta interface {
	Name() string
	Description() []string
	RequiredVars() []string
	GetValue(key string) string
	OpenShiftVersion() string
}

type DefaultAddOnMeta struct {
	headers map[string]interface{}
}

func NewAddOnMeta(headers map[string]interface{}) (AddOnMeta, error) {
	for _, tag := range requiredMetaTags {
		if headers[tag] == nil {
			return nil, errors.New(fmt.Sprintf("Metadata does not contain an mandatory entry for '%s'", tag))
		}
		switch v := headers[tag].(type) {
		case string:
			if v == "" {
				return nil, errors.New(fmt.Sprintf("Metadata does not contain an mandatory entry for '%s'", tag))
			}
		case []string:
			if len(v) == 0 {
				return nil, errors.New(fmt.Sprintf("Metadata does not contain an mandatory entry for '%s'", tag))
			}
		default:
			return nil, nil
		}
	}
	if headers[RequiredOpenShiftVersion] != nil {
		if !checkVersionSemantic(headers[RequiredOpenShiftVersion].(string)) {
			return nil, errors.New("Add-on only support OpenShift version semantics eg. 3.6.0 or >3.6.0, <3.9.0 or >=3.5 etc.")
		}
	}

	metaData := &DefaultAddOnMeta{headers: headers}
	return metaData, nil
}

func (meta *DefaultAddOnMeta) String() string {
	return fmt.Sprintf("%#v", meta)
}

func (meta *DefaultAddOnMeta) Name() string {
	return meta.headers[requiredMetaTags[0]].(string)
}

func (meta *DefaultAddOnMeta) Description() []string {
	return meta.headers[requiredMetaTags[1]].([]string)
}

func (meta *DefaultAddOnMeta) RequiredVars() []string {
	if val, contains := meta.headers[requiredVars].(string); contains {
		return meta.splitAndTrim(val)
	} else {
		return []string{}
	}
}

func (meta *DefaultAddOnMeta) GetValue(key string) string {
	return meta.headers[key].(string)
}

func (meta *DefaultAddOnMeta) OpenShiftVersion() string {
	if val, contains := meta.headers[RequiredOpenShiftVersion].(string); contains {
		return val
	}
	return anyOpenShiftVersion
}

func (meta *DefaultAddOnMeta) splitAndTrim(s string) []string {
	// Trims the stream and then splits
	trimmed := strings.TrimSpace(s)
	split := strings.Split(trimmed, ",")
	cleanSplit := make([]string, len(split))
	for i, val := range split {
		cleanSplit[i] = strings.TrimSpace(val)
	}
	return cleanSplit
}

func checkVersionSemantic(version string) bool {
	// Strict match for <major> or <major>.<minor> or <major>.<minor>.<patch>
	// (>=|>|<|<=)3.6.0, (>=|>|<|<=)3.6.0
	match, _ := regexp.MatchString("^(|>|>=|<|<=)[0-9]+(.[0-9]+){0,2}(|\\s*,\\s*(|>|>=|<|<=)[0-9]+(.[0-9]+){0,2})$", version)
	return match
}
