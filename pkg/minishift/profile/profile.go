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
	} else {
		profileList = append(profileList, constants.DefaultProfileName)
	}

	if !filehelper.IsDirectory(profileBaseDir) {
		return profileList
	}
	files, err := ioutil.ReadDir(profileBaseDir)
	if err != nil {
		return profileList
	}

	for _, f := range files {
		profileList = append(profileList, f.Name())
	}
	return profileList
}

//Set Active Profile and also it makes sure that we have one
// active profile at one point of time.
func SetActiveProfile(name string) error {
	activeProfile := config.AllInstancesConfig.ActiveProfile
	if name != activeProfile {
		config.AllInstancesConfig.ActiveProfile = name
	}
	err := config.AllInstancesConfig.Write()
	if err != nil {
		return fmt.Errorf("Error updating profile information in config. %s", err)
	}
	return nil
}

// Get Active Profile from AllInstancesConfig
func GetActiveProfile() string {
	return config.AllInstancesConfig.ActiveProfile
}

// Placeholder function to change constants related to a VM instance
// This needs a better solution than this as these constats should not be
// touched outside of cmd/root.go. However cluster.GetHostStatus(api) uses
// constants.MachineName inside the function.
// This function should not be used outside profile set, list and delete
func UpdateMiniConstants(profileName string) {
	constants.ProfileName = profileName
	constants.MachineName = constants.ProfileName
	constants.Minipath = constants.GetProfileHomeDir()
}
