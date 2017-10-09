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
	minishiftStrings "github.com/minishift/minishift/pkg/util/strings"
	"regexp"
)

var requiredMetaTags = []string{NameMetaTagName, DescriptionMetaTagName}

const (
	requiredVars             = "Required-Vars"
	NameMetaTagName          = "Name"
	DescriptionMetaTagName   = "Description"
	RequiredOpenShiftVersion = "OpenShift-Version"
	anyOpenShiftVersion      = ""
	varDefaults              = "Var-Defaults"
)

type RequiredVar struct {
	Key   string
	Value string
}

// AddOnMeta defines a set of meta data for an AddOn. Name and Description are required. Others are optional.
type AddOnMeta interface {
	Name() string
	Description() []string
	RequiredVars() ([]string, error)
	VarDefaults() ([]RequiredVar, error)
	GetValue(key string) string
	OpenShiftVersion() string
}

type DefaultAddOnMeta struct {
	headers map[string]interface{}
}

func NewAddOnMeta(headers map[string]interface{}) (AddOnMeta, error) {
	if err := requiredMetaTagsCheck(headers); err != nil {
		return nil, err
	}
	if err := requiredOpenShiftVersionCheck(headers); err != nil {
		return nil, err
	}
	if err := varDefaultsCheck(headers); err != nil {
		return nil, err
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

func (meta *DefaultAddOnMeta) RequiredVars() ([]string, error) {
	if val, contains := meta.headers[requiredVars].(string); contains {
		return minishiftStrings.SplitAndTrim(val, ",")
	} else {
		return []string{}, nil
	}
}

func (meta *DefaultAddOnMeta) VarDefaults() ([]RequiredVar, error) {
	// Ignore errors as checking has been done during varDefaultsCheck as part of NewAddOnMeta
	if val, contains := meta.headers[varDefaults].(string); contains {
		items, _ := minishiftStrings.SplitAndTrim(val, ",")
		defaults := make([]RequiredVar, 0, len(items))
		for _, item := range items {
			varDefault, _ := minishiftStrings.SplitAndTrim(item, "=")
			defaults = append(defaults, RequiredVar{Key: varDefault[0], Value: varDefault[1]})
		}

		return defaults, nil
	}

	return []RequiredVar{}, nil
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

func checkVersionSemantic(version string) bool {
	// Strict match for <major> or <major>.<minor> or <major>.<minor>.<patch>
	// (>=|>|<|<=)3.6.0, (>=|>|<|<=)3.6.0
	match, _ := regexp.MatchString("^(|>|>=|<|<=)[0-9]+(.[0-9]+){0,2}(|\\s*,\\s*(|>|>=|<|<=)[0-9]+(.[0-9]+){0,2})$", version)
	return match
}

func requiredMetaTagsCheck(headers map[string]interface{}) error {
	for _, tag := range requiredMetaTags {
		if headers[tag] == nil {
			return errors.New(fmt.Sprintf("Metadata does not contain a mandatory entry for '%s'", tag))
		}
		switch v := headers[tag].(type) {
		case string:
			if v == "" {
				return errors.New(fmt.Sprintf("Metadata does not contain a mandatory entry for '%s'", tag))
			}
		case []string:
			if len(v) == 0 {
				return errors.New(fmt.Sprintf("Metadata does not contain a mandatory entry for '%s'", tag))
			}
		default:
			return nil
		}
	}
	return nil
}

func requiredOpenShiftVersionCheck(headers map[string]interface{}) error {
	if headers[RequiredOpenShiftVersion] != nil {
		if !checkVersionSemantic(headers[RequiredOpenShiftVersion].(string)) {
			return errors.New("Add-on only support OpenShift version semantics eg. 3.6.0 or >3.6.0, <3.9.0 or >=3.5 etc.")
		}
	}

	return nil
}

func varDefaultsCheck(headers map[string]interface{}) error {
	if val, contains := headers[varDefaults].(string); contains {
		items, err := minishiftStrings.SplitAndTrim(val, ",")
		if err != nil {
			return err
		}

		for _, item := range items {
			varDefault, err := minishiftStrings.SplitAndTrim(item, "=")
			if err != nil {
				return err
			}
			// Expect varDefault to be only having key and value
			if len(varDefault) != 2 {
				return errors.New(fmt.Sprintf("'%s' is not well formed", val))
			}

			if varDefault[0] == "" || varDefault[1] == "" {
				return errors.New(fmt.Sprintf("'%s' is not well formed", val))
			}
		}
	}

	return nil
}
