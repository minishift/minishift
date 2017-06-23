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

	"github.com/minishift/minishift/pkg/minishift/config"
)

//Add profile information to allinstancesconfig
func AddProfileToConfig(name string) error {
	profile := newProfile(name)
	config.AllInstancesConfig.Profiles = append(config.AllInstancesConfig.Profiles, profile)
	err := config.AllInstancesConfig.Write()
	if err != nil {
		return fmt.Errorf("Error adding profile information. %s", err)
	}
	return nil
}

func newProfile(name string) config.Profile {
	return config.Profile{
		Name:  name,
		InUse: false,
	}
}

//Returns the list of profile names
func GetProfileNameList() []string {
	var profileList []string
	profiles := config.AllInstancesConfig.Profiles
	for i := range profiles {
		profile := profiles[i]
		profileList = append(profileList, profile.Name)
	}
	return profileList
}

//RemoveProfile
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
			profiles[i].InUse = true
		} else {
			profiles[i].InUse = false
		}
	}

	config.AllInstancesConfig.Profiles = profiles
	err := config.AllInstancesConfig.Write()
	if err != nil {
		return fmt.Errorf("Error updating profile information in config. %s", err)
	}
	return nil
}

//Unset Active Profile and also it makes sure that we have one
// active profile at one point of time.
func ResetActiveProfile() error {
	profiles := config.AllInstancesConfig.Profiles

	if GetActiveProfile() == "" {
		return nil
	}

	for i := range profiles {
		profiles[i].InUse = false
	}

	config.AllInstancesConfig.Profiles = profiles
	err := config.AllInstancesConfig.Write()
	if err != nil {
		return fmt.Errorf("Error updating profile information in config. %s", err)
	}
	return nil
}

//Get Active Profile from AllInstancesConfig
func GetActiveProfile() string {
	profiles := config.AllInstancesConfig.Profiles
	for i := range profiles {
		profile := profiles[i]
		if profile.InUse == true {
			return profile.Name
		}
	}
	return ""
}
