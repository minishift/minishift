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
	"errors"
	"os"
	"regexp"
)

const (
	booleanTruthyMatch = "^(?i:true|y|yes|1|on)$"
	booleanFalsyMatch  = "^(?i:false|n|no|0|off)$"
)

var (
	BooleanFormatError  = errors.New("Invalid format for boolean value.")
	BooleanNoValueError = errors.New("No value given.")
)

// GetBoolEnv returns truthy or falsy value of given environmental variable.
// It returns an error on failure with a default of false.
func GetBoolEnv(varName string) (bool, error) {
	varValue := os.Getenv(varName)

	// no value given
	if varValue == "" {
		return false, BooleanNoValueError
	}

	truthy := regexp.MustCompile(booleanTruthyMatch)
	falsy := regexp.MustCompile(booleanFalsyMatch)

	if truthy.FindString(varValue) != "" {
		return true, nil
	}
	if falsy.FindString(varValue) != "" {
		return false, nil
	}

	return false, BooleanFormatError
}
