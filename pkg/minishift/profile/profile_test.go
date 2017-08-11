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

	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
)

var (
	testDir        string
	configFilePath string
)

func TestAdditionOfProfileToConfig(t *testing.T) {
	setup(t)
	defer teardown()
	profileName := "foo"

	configFilePath = filepath.Join(testDir, "fake-machine.json")
	minishiftConfig.AllInstancesConfig, _ = minishiftConfig.NewAllInstancesConfig(configFilePath)

	AddProfileToConfig(profileName)
	profileList := GetProfileNameList()
	if len(profileList) > 1 {
		t.Errorf("Number of profiles is more than expected")
	}
	if profileName != profileList[0] {
		t.Errorf("Expected profile name %s but actual found %s", profileName, profileList[0])
	}
}

func TestRemovableOfProfileToConfig(t *testing.T) {
	setup(t)
	defer teardown()
	profileName := "foo"

	AddProfileToConfig(profileName)
	RemoveProfile(profileName)
	profileList := GetProfileNameList()
	if len(profileList) > 0 {
		t.Errorf("There was no profile expected")
	}
}

func TestSetingActiveProfile(t *testing.T) {
	setup(t)
	defer teardown()
	profileName := "foo"

	AddProfileToConfig(profileName)
	SetActiveProfile(profileName)
	actualProfileName := GetActiveProfile()
	if actualProfileName != profileName {
		t.Errorf("Expected profile name: %s does not match actual profile name: %s", profileName, actualProfileName)
	}
}

//TBD: Will add some more unit tests

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
