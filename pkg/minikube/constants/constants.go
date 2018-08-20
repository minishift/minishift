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

	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/version"
)

const (
	APIServerPort                    = 8443             // Port that the API server should listen on
	MiniShiftEnvPrefix               = "MINISHIFT"      // Prefix for the environmental variables
	MiniShiftHomeEnv                 = "MINISHIFT_HOME" // Environment variable used to change the Minishift home directory
	VersionPrefix                    = "v"
	MinimumSupportedOpenShiftVersion = "v3.9.0"
	RefactoredOcVersion              = "v3.10.0-alpha.0" // From v3.10.0 oc binary is refactored and don't have lot of options by default.
	DefaultMemory                    = "4GB"
	DefaultCPUS                      = 2
	DefaultDiskSize                  = "20GB"
	UpdateMarkerFileName             = "updated"
	DefaultMachineName               = "minishift"
	DefaultProfileName               = "minishift"
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

	DefaultB2dIsoUrl    = "https://github.com/minishift/minishift-b2d-iso/releases/download/" + version.GetB2dIsoVersion() + "/" + "minishift-b2d.iso"
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
		return filepath.Join(util.HomeDir(), ".minishift", "profiles", profile)
	}
	return GetMinishiftHomeDir()
}

// getMinishiftHomeDir determines the Minishift home directory where all state information is kept.
// The default directory is .minishift in the users HOME directory which can be overwritten by the MINISHIFT_HOME
// environment variable
func GetMinishiftHomeDir() string {
	homeEnv, ok := os.LookupEnv(MiniShiftHomeEnv)
	if ok {
		return homeEnv
	} else {
		return filepath.Join(util.HomeDir(), ".minishift")
	}
}
