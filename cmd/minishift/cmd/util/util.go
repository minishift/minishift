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

package util

import (
	"fmt"
	"regexp"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/os/atexit"

	"github.com/minishift/minishift/out/bindata"
)

var (
	DefaultAssets = []string{"anyuid", "admin-user", "xpaas", "registry-route"}
)

func VMExists(client *libmachine.Client, machineName string) bool {
	exists, err := client.Exists(machineName)
	if err != nil {
		atexit.ExitWithMessage(1, "Cannot determine the state of Minishift VM.")
	}
	return exists
}

func ExitIfUndefined(client *libmachine.Client, machineName string) {
	exists := VMExists(client, machineName)
	if !exists {
		atexit.ExitWithMessage(0, fmt.Sprintf("Running this command requires an existing '%s' VM, but no VM is defined.", machineName))
	}
}

func IsHostRunning(driver drivers.Driver) bool {
	return drivers.MachineInState(driver, state.Running)()
}

func IsHostStopped(driver drivers.Driver) bool {
	return drivers.MachineInState(driver, state.Stopped)()
}

func ExitIfNotRunning(driver drivers.Driver, machineName string) {
	running := IsHostRunning(driver)
	if !running {
		atexit.ExitWithMessage(0, fmt.Sprintf("Running this command requires a running '%s' VM, but no VM is running.", machineName))
	}
}

func GetVMStatus(profileName string) string {
	var status string
	profileActions.UpdateProfileConstants(profileName)
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()
	status, err := cluster.GetHostStatus(api)
	if err != nil {
		status = fmt.Sprintf("Error getting the VM status: %s", err.Error())
	}
	return status
}

func DoesVMExist(profileName string) bool {
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()
	return VMExists(api, profileName)
}

// UnpackAddons will unpack the default addons into addons default dir
func UnpackAddons(addonsDir string) error {
	for _, asset := range DefaultAssets {
		if err := bindata.RestoreAssets(addonsDir, asset); err != nil {
			return err
		}
	}

	return nil
}

// IsValidProfile return true if a given profile exist
func IsValidProfile(profileName string) bool {
	profileList := profileActions.GetProfileList()
	for i := range profileList {
		if profileList[i] == profileName {
			return true
		}
	}
	return false
}

// IsValidProfileName return true if profile name follow ^[a-zA-Z0-9]+[\w+-]*$
func IsValidProfileName(profileName string) bool {
	rr := regexp.MustCompile(`^[a-zA-Z0-9]+[\w+-]*$`)
	return rr.MatchString(profileName)
}
