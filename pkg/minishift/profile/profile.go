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

	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/config"
)

// Add profile information to allinstancesconfig
func AddProfileToConfig(name string) error {
	var doesProfileExist bool
	existingProfiles := config.AllInstancesConfig.Profiles
	for i := range existingProfiles {
		if name == existingProfiles[i].Name {
			doesProfileExist = true
			return nil
		}
	}
	if !doesProfileExist {
		profile := newProfile(name)
		config.AllInstancesConfig.Profiles = append(config.AllInstancesConfig.Profiles, profile)
		err := config.AllInstancesConfig.Write()
		if err != nil {
			return fmt.Errorf("Error adding profile information. %s", err)
		}
	}
	return nil
}

func newProfile(name string) config.Profile {
	return config.Profile{
		Name:   name,
		Active: false,
	}
}

// Returns the list of profile names
func GetProfileNameList() []string {
	var profileList []string
	profiles := config.AllInstancesConfig.Profiles
	for i := range profiles {
		profile := profiles[i]
		profileList = append(profileList, profile.Name)
	}
	return profileList
}

// Returns a map of profile name and if it is Active
func GetProfileMap() map[string]bool {
	profileMap := make(map[string]bool)
	profiles := config.AllInstancesConfig.Profiles
	for i := range profiles {
		profileMap[profiles[i].Name] = profiles[i].Active
	}
	return profileMap
}

func RemoveProfile(name string) error {
	profiles := config.AllInstancesConfig.Profiles
	for i := range profiles {
		profile := profiles[i]
		if profile.Name == name {
			profiles = append(profiles[:i], profiles[i+1:]...)
			break
		}
	}
	config.AllInstancesConfig.Profiles = profiles
	err := config.AllInstancesConfig.Write()
	if err != nil {
		return fmt.Errorf("Error removing profile information from config. %s", err)
	}
	return nil
}

//Set Active Profile and also it makes sure that we have one
// active profile at one point of time.
func SetActiveProfile(name string) error {
	profiles := config.AllInstancesConfig.Profiles
	for i := range profiles {
		if profiles[i].Name == name {
			profiles[i].Active = true
		} else {
			profiles[i].Active = false
		}
	}

	config.AllInstancesConfig.Profiles = profiles
	err := config.AllInstancesConfig.Write()
	if err != nil {
		return fmt.Errorf("Error updating profile information in config. %s", err)
	}
	return nil
}

// Get Active Profile from AllInstancesConfig
func GetActiveProfile() string {
	profiles := config.AllInstancesConfig.Profiles
	for i := range profiles {
		profile := profiles[i]
		if profile.Active == true {
			return profile.Name
		}
	}
	return ""
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
