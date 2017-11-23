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

package profile

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/stretchr/testify/assert"
)

var (
	testDir        string
	configFilePath string
)

func TestSetingActiveProfile(t *testing.T) {
	setup(t)
	defer teardown()

	profileName := "foo"
	SetActiveProfile(profileName)
	actualProfileName := GetActiveProfile()
	assert.Equal(t, profileName, actualProfileName, "Profile name doesn't match")
}

func TestSetingGetProfileList(t *testing.T) {
	setup(t)
	defer teardown()

	miniPath, err := ioutil.TempDir("", "minishift-test-profile-")
	assert.NoError(t, err)
	os.Setenv("MINISHIFT_HOME", miniPath)
	defer os.Unsetenv("MINISHIFT_HOME")

	profileList := []string{constants.DefaultProfileName, "abc", "xyz"}

	for _, profile := range profileList {
		if profile != constants.DefaultProfileName {
			dirPath := filepath.Join(miniPath, "profiles", profile)
			err := os.MkdirAll(dirPath, os.ModePerm)
			assert.NoError(t, err, "Error creating profiles directory")
		}
	}
	actualProfileList := GetProfileList()
	assert.EqualValues(t, profileList, actualProfileList, "Profile lists do not match")
}

func setup(t *testing.T) {
	var err error
	testDir, err = ioutil.TempDir("", "minishift-test-profile-")

	if err != nil {
		t.Error(err)
	}
	configFilePath = filepath.Join(testDir, "fake-machine.json")
	minishiftConfig.AllInstancesConfig, _ = minishiftConfig.NewAllInstancesConfig(configFilePath)
}

func teardown() {
	os.RemoveAll(testDir)
}
