/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package constants

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/minishift/minishift/pkg/version"
)

const (
	APIServerPort                    = 8443             // Port that the API server should listen on
	MiniShiftEnvPrefix               = "MINISHIFT"      // Prefix for the environmental variables
	MiniShiftHomeEnv                 = "MINISHIFT_HOME" // Environment variable used to change the Minishift home directory
	VersionPrefix                    = "v"
	MinimumSupportedOpenShiftVersion = "v3.10.0"
	DefaultMemory                    = "4GB"
	DefaultCPUS                      = 2
	DefaultDiskSize                  = "20GB"
	UpdateMarkerFileName             = "updated"
	DefaultMachineName               = "minishift"
	DefaultProfileName               = "minishift"
	DefaultTimeZone                  = "UTC" // This is what we have in our ISO kickstart template.
)

var (
	// Fix for windows
	Minipath    = GetMinishiftHomeDir()
	ProfileName = DefaultProfileName
	MachineName = DefaultMachineName // Name to use for the VM

	KubeConfigPath        = filepath.Join(Minipath, "machines", MachineName+"_kubeconfig")
	ConfigFile            = MakeMiniPath("config", "config.json")
	GlobalConfigFile      = filepath.Join(Minipath, "config", "global.json")
	AllInstanceConfigPath = filepath.Join(Minipath, "config", "allinstances.json")

	DefaultCentOsIsoUrl = "https://github.com/minishift/minishift-centos-iso/releases/download/" + version.GetCentOsIsoVersion() + "/" + "minishift-centos7.iso"
)

// MakeMiniPath is a utility to calculate a relative path to our directory.
func MakeMiniPath(fileName ...string) string {
	args := []string{Minipath}
	args = append(args, fileName...)
	return filepath.Join(args...)
}

// GetProfileHomeDir determines the base directory for the specified profile name
func GetProfileHomeDir(profile string) string {
	if profile != DefaultProfileName {
		homeEnv, ok := os.LookupEnv(MiniShiftHomeEnv)
		if ok {
			return filepath.Join(homeEnv, "profiles", profile)
		}
		return filepath.Join(GetHomeDir(), ".minishift", "profiles", profile)
	}
	return GetMinishiftHomeDir()
}

func GetProfileConfigFile(profile string) string {
	return filepath.Join(GetProfileHomeDir(profile), "config", "config.json")
}

// getMinishiftHomeDir determines the Minishift home directory where all state information is kept.
// The default directory is .minishift in the users HOME directory which can be overwritten by the MINISHIFT_HOME
// environment variable
func GetMinishiftHomeDir() string {
	homeEnv, ok := os.LookupEnv(MiniShiftHomeEnv)
	if ok {
		return homeEnv
	} else {
		return filepath.Join(GetHomeDir(), ".minishift")
	}
}

// GetMinishiftProfilesDir returns the path MINISHIFT_HOME/profiles
func GetMinishiftProfilesDir() string {
	minishiftHomeDir := GetMinishiftHomeDir()
	return filepath.Join(minishiftHomeDir, "profiles")
}

// GetHomeDir returns the home directory for the current user
func GetHomeDir() string {
	if runtime.GOOS == "windows" {
		if homeDrive, homePath := os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"); len(homeDrive) > 0 && len(homePath) > 0 {
			homeDir := filepath.Join(homeDrive, homePath)
			if _, err := os.Stat(homeDir); err == nil {
				return homeDir
			}
		}
		if userProfile := os.Getenv("USERPROFILE"); len(userProfile) > 0 {
			if _, err := os.Stat(userProfile); err == nil {
				return userProfile
			}
		}
	}
	return os.Getenv("HOME")
}
