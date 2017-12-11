// +build integration

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
	"fmt"
	"regexp"
	"strings"
)

func CompareExpectedWithActualContains(expected string, actual string) error {
	if !strings.Contains(actual, expected) {
		return fmt.Errorf("Output did not match. Expected: '%s', Actual: '%s'", expected, actual)
	}

	return nil
}

func CompareExpectedWithActualNotContains(notexpected string, actual string) error {
	if strings.Contains(actual, notexpected) {
		return fmt.Errorf("Output did match. Not expected: '%s', Actual: '%s'", notexpected, actual)
	}

	return nil
}

func CompareExpectedWithActualEquals(expected string, actual string) error {
	if actual != expected {
		return fmt.Errorf("Output did not match. Expected: '%s', Actual: '%s'", expected, actual)
	}

	return nil
}

func CompareExpectedWithActualNotEquals(notexpected string, actual string) error {
	if actual == notexpected {
		return fmt.Errorf("Output did match. Not expected: '%s', Actual: '%s'", notexpected, actual)
	}

	return nil
}

func PerformRegexMatch(regex string, input string) (bool, error) {
	compRegex, err := regexp.Compile(regex)
	if err != nil {
		return false, fmt.Errorf("Expected value must be a valid regular expression statement: ", err)
	}

	return compRegex.MatchString(input), nil
}

func CompareExpectedWithActualMatchesRegex(expected string, actual string) error {
	matches, err := PerformRegexMatch(expected, actual)
	if err != nil {
		return err
	} else if !matches {
		return fmt.Errorf("Output did not match. Expected: '%s', Actual: '%s'", expected, actual)
	}

	return nil
}

func CompareExpectedWithActualNotMatchesRegex(notexpected string, actual string) error {
	matches, err := PerformRegexMatch(notexpected, actual)
	if err != nil {
		return err
	} else if matches {
		return fmt.Errorf("Output did match. Not expected: '%s', Actual: '%s'", notexpected, actual)
	}

	return nil
}
