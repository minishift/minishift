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

package clusterup

import (
	"os"
	"testing"

	utilStrings "github.com/minishift/minishift/pkg/util/strings"
	"github.com/stretchr/testify/assert"
)

func Test_DetermineOcVersion(t *testing.T) {
	var versionTests = []struct {
		inputVersion string
		outVersion   string
	}{
		{"v1.5.0", "v3.6.1"},
		{"v3.6.0", "v3.6.1"},
		{"v3.7.0", "v3.7.0"},
	}
	for _, version := range versionTests {
		actualOcVersion := DetermineOcVersion(version.inputVersion)
		assert.Equal(t, version.outVersion, actualOcVersion)
	}
}

func Test_invalid_addon_variable_leads_to_error_in_context_creation(t *testing.T) {
	context, err := GetExecutionContext("127.0.0.1", "foo.bar", []string{"FOOBAR"}, nil, nil)
	assert.Error(t, err, "There should have been an error due to incorrect addon env variable.")
	assert.Nil(t, context, "There should be no InterpolationContext returned.")
}

func Test_addon_variable_can_be_interpolated(t *testing.T) {
	assertInterpolation([]string{"FOO=env.BAR"}, "#{FOO}", "#{FOO}", t)
}

func Test_nil_can_be_passed_to_create_context(t *testing.T) {
	_, err := GetExecutionContext("127.0.0.1", "foo.bar", nil, nil, nil)

	assert.NoError(t, err, "Error in getting execution context")
}

func Test_addon_variable_can_be_interpolated_from_environment(t *testing.T) {
	env := os.Environ()
	os.Clearenv()
	defer resetEnv(env)

	assertInterpolation([]string{"FOO=env.BAR"}, "#{FOO}", "#{FOO}", t)

	os.Setenv("BAR", "SNAFU")
	assertInterpolation([]string{"FOO=env.BAR"}, "#{FOO}", "SNAFU", t)

	os.Unsetenv("BAR")
	assertInterpolation([]string{"FOO=env.BAR"}, "#{FOO}", "#{FOO}", t)
}

func assertInterpolation(variables []string, testString string, expectedResult string, t *testing.T) {
	context, err := GetExecutionContext("127.0.0.1", "foo.bar", variables, nil, nil)
	assert.NoError(t, err)

	result := context.Interpolate(testString)
	assert.Equal(t, expectedResult, result)
}

func resetEnv(env []string) {
	os.Clearenv()
	for _, envSetting := range env {
		envTokens, _ := utilStrings.SplitAndTrim(envSetting, "=")
		os.Setenv(envTokens[0], envTokens[1])
	}
}
