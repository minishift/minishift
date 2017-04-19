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

// MachineName is the name to use for the VM.
const MachineName = "minishift"

// APIServerPort is the port that the API server should listen on.
const APIServerPort = 8443

// Fix for windows
var Minipath = getMinishiftHomeDir()

// MiniShiftEnvPrefix is the prefix for the environmental variables
const MiniShiftEnvPrefix = "MINISHIFT"

// MiniShiftHomeEnv is the environment variable used to change the Minishift home directory
const MiniShiftHomeEnv = "MINISHIFT_HOME"

const VersionPrefix = "v"

// Minimum Openshift supported version
const MinOpenshiftSuportedVersion = "v1.4.1"

const (
	DefaultMemory   = 2048
	DefaultCPUS     = 2
	DefaultDiskSize = "20g"
)

var KubeConfigPath = filepath.Join(Minipath, "machines", MachineName+"_kubeconfig")

var isoVersion = version.GetIsoVersion()
var isoName = "minishift-b2d.iso"
var githubProject = "https://github.com/minishift/minishift-b2d-iso"

var DefaultIsoUrl = githubProject + "/releases/download/" + isoVersion + "/" + isoName

var ConfigFile = MakeMiniPath("config", "config.json")

// MakeMiniPath is a utility to calculate a relative path to our directory.
func MakeMiniPath(fileName ...string) string {
	args := []string{Minipath}
	args = append(args, fileName...)
	return filepath.Join(args...)
}

// getMinishiftHomeDir determines the Minishift home directory where all state information is kept.
// The default directory is .minishift in the users HOME directory which can be overwritten by the MINISHIFT_HOME
// environment variable
func getMinishiftHomeDir() string {
	homeEnv, ok := os.LookupEnv(MiniShiftHomeEnv)
	if ok {
		return homeEnv
	} else {
		return filepath.Join(util.HomeDir(), ".minishift")
	}
}
