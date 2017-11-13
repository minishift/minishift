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

	"reflect"

	"fmt"

	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
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
	if actualProfileName != profileName {
		t.Errorf("Expected profile name: %s does not match actual profile name: %s", profileName, actualProfileName)
	}
}

func TestSetingGetProfileList(t *testing.T) {
	setup(t)
	defer teardown()

	miniPath, err := ioutil.TempDir("", "minishift-test-profile-")
	if err != nil {
		t.Error(err)
	}
	os.Setenv("MINISHIFT_HOME", miniPath)
	defer os.Unsetenv("MINISHIFT_HOME")

	profileList := []string{constants.DefaultProfileName, "abc", "xyz"}

	for _, profile := range profileList {
		if profile != constants.DefaultProfileName {
			dirPath := filepath.Join(miniPath, "profiles", profile)
			err := os.MkdirAll(dirPath, os.ModePerm)
			if err != nil {
				fmt.Println(fmt.Sprintf("%s", err.Error()))
			}
		}
	}
	actualProfileList := GetProfileList()
	if !reflect.DeepEqual(profileList, actualProfileList) {
		t.Error("Expected profile name does not match actual profile name")
	}
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
