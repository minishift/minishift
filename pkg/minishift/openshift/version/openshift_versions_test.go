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
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/github"
	"github.com/minishift/minishift/pkg/version"
	"github.com/stretchr/testify/assert"
)

func TestPrintUpStreamVersions(t *testing.T) {
	EnsureGitHubApiAccessTokenSet(t)
	testDir, err := ioutil.TempDir("", "minishift-config-")
	assert.NoError(t, err)
	defer os.RemoveAll(testDir)

	f, err := os.Create(testDir + "out.txt")
	assert.NoError(t, err)
	defer f.Close()

	os.Stdout = f
	defaultVersion := version.GetOpenShiftVersion()
	err = PrintUpStreamVersions(f, constants.MinimumSupportedOpenShiftVersion, defaultVersion)
	assert.NoError(t, err)

	_, err = f.Seek(0, 0)
	assert.NoError(t, err, "Error setting offset back")
	data, err := ioutil.ReadAll(f)
	assert.NoError(t, err)
	actualStdout := string(data)
	assert.Contains(t, actualStdout, constants.MinimumSupportedOpenShiftVersion)
	assert.Contains(t, actualStdout, defaultVersion)
}

func EnsureGitHubApiAccessTokenSet(t *testing.T) {
	if github.GetGitHubApiToken() == "" {
		t.Skip("Skipping GitHub API based test, because no access token is defined in the environment.\n " +
			"To run this test check https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/ and set for example MINISHIFT_GITHUB_API_TOKEN (see github.go).")
	}
}

func TestPrintDownStreamVersions(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-config-")
	assert.NoError(t, err)
	defer os.RemoveAll(testDir)

	f, err := os.Create(testDir + "out.txt")
	assert.NoError(t, err)
	defer f.Close()

	os.Stdout = f
	err = PrintDownStreamVersions(f, "v3.4.1.10")
	assert.NoError(t, err)
	_, err = f.Seek(0, 0)
	assert.NoError(t, err, "Error setting offset back")

	data, err := ioutil.ReadAll(f)
	assert.NoError(t, err, "Error reading file")

	actualStdout := string(data)
	assert.Contains(t, actualStdout, "v3.4.1.10")
}

func TestIsGreaterOrEqualToBaseVersion(t *testing.T) {
	var versionTestData = []struct {
		openshiftVersion     string
		baseOpenshiftVersion string
		expectedResult       bool
		expectedErr          error
	}{
		{"v3.6.0", "v3.7.0", false, nil},
		{"v3.6.0-alpha.1", "v3.6.0-alpha.2", false, nil},
		{"v3.7.1", "v3.7.1", true, nil},
		{"v3.6.0-rc1", "v3.6.0-alpha.1", true, nil},
		{"v3.6.0-alpha.1", "v3.6.0-beta.0", false, nil},
		{"v1.4.1", "v1.4.1", true, nil},
		{"v1.5.0-alpha.0", "v1.4.1", true, nil},
		{"v1.5.1-beta.0", "v1.4.1", true, nil},
		{"foo", "v1.4.1", false, errors.New("Invalid version format 'foo': No Major.Minor.Patch elements found")},
		{"151", "v1.4.1", false, errors.New("Invalid version format '151': No Major.Minor.Patch elements found")},
	}

	for _, versionTest := range versionTestData {
		actualResult, actualErr := IsGreaterOrEqualToBaseVersion(versionTest.openshiftVersion, versionTest.baseOpenshiftVersion)

		assert.Equal(t, versionTest.expectedResult, actualResult)

		if actualErr != nil {
			assert.EqualError(t, actualErr, versionTest.expectedErr.Error())
		}
	}
}

func TestIsPrerelease(t *testing.T) {
	var versionTestData = []struct {
		openshiftVersion string
		expectedResult   bool
	}{
		{"v3.6.0", false},
		{"v3.6.0-alpha.1", true},
		{"v3.9.0-alpha.3", true},
		{"v3.8.0-rc.1", true},
		{"v3.6.0-beta", true},
	}

	for _, versionTest := range versionTestData {
		actualResult := isPrerelease(versionTest.openshiftVersion)
		assert.Equal(t, versionTest.expectedResult, actualResult)
	}
}

func TestOpenShiftTagsByAscending(t *testing.T) {
	var testTags = []string{"v3.7.0", "v3.10.0", "v3.7.2", "v3.9.0", "v3.7.1"}

	sortTags, _ := OpenShiftTagsByAscending(testTags, "v3.7.0", "v3.9.0")
	assert.Equal(t, sortTags[len(sortTags)-1], "v3.10.0")
}
