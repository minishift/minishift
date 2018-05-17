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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util/filehelper"
)

// Returns the list of profile names
func GetProfileList() []string {
	var profileList []string
	baseDir := constants.GetMinishiftHomeDir()
	profileBaseDir := filepath.Join(baseDir, "profiles")

	if !filehelper.IsDirectory(baseDir) {
		return profileList
	}
	profileList = append(profileList, constants.DefaultProfileName)

	if !filehelper.IsDirectory(profileBaseDir) {
		return profileList
	}

	files, err := ioutil.ReadDir(profileBaseDir)
	if err != nil {
		return profileList
	}

	for _, f := range files {
		// Skip non-directory and hidden stuffs
		match, _ := regexp.MatchString("^\\.", f.Name())
		if !filehelper.IsDirectory(filepath.Join(profileBaseDir, f.Name())) || match {
			continue
		}
		profileList = append(profileList, f.Name())
	}
	return profileList
}

// Set Active Profile and also it makes sure that we have one
// active profile at one point of time.
func SetActiveProfile(name string) error {
	activeProfile := config.AllInstancesConfig.ActiveProfile
	if name != activeProfile {
		config.AllInstancesConfig.ActiveProfile = name
	}
	err := config.AllInstancesConfig.Write()
	if err != nil {
		return fmt.Errorf("Error updating active profile information for '%s' in config. %s", name, err)
	}
	return nil
}

// Get Active Profile from AllInstancesConfig
func GetActiveProfile() string {
	if config.AllInstancesConfig != nil {
		return config.AllInstancesConfig.ActiveProfile
	}
	return ""
}

// Placeholder function to change variables related to a VM instance
// This needs a better solution than this as these variables should not be
// changed outside of cmd/root.go. However cluster.GetHostStatus(api) uses
// constants.MachineName inside the function.
// This is a temporary fix and we will findout a better way to do it.
func UpdateProfileConstants(profileName string) {
	constants.ProfileName = profileName
	constants.MachineName = constants.ProfileName
	constants.Minipath = constants.GetProfileHomeDir(constants.ProfileName)
}

func SetDefaultProfileActive() error {
	err := SetActiveProfile(constants.DefaultProfileName)
	if err != nil {
		return fmt.Errorf("Error while setting default profile '%s' active: %s", constants.DefaultProfileName, err.Error())
	}
	return nil
}
