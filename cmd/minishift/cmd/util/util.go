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
	"strings"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/os/atexit"

	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	cmdState "github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/out/bindata"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/util/shell"
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
	profileDirs := cmdState.NewMinishiftDirs(constants.GetProfileHomeDir(profileName))

	api := libmachine.NewClient(profileDirs.Home, profileDirs.Certs)
	defer api.Close()
	status, err := cluster.GetHostStatus(api, profileName)
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
	rr := regexp.MustCompile(`^[a-zA-Z0-9]+[a-zA-Z0-9-]*$`)
	return rr.MatchString(profileName)
}

func GetNoProxyConfig(api libmachine.API) (string, string, error) {
	host, err := api.Load(constants.MachineName)
	if err != nil {
		return "", "", fmt.Errorf("Error getting IP: %s", err)
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return "", "", fmt.Errorf("Error getting host IP: %s", err)
	}

	noProxyVar, noProxyValue := shell.FindNoProxyFromEnv()

	// Add the minishift VM to the no_proxy list idempotently.
	switch {
	case noProxyValue == "":
		noProxyValue = ip
	case strings.Contains(noProxyValue, ip):
	// IP already in no_proxy list, nothing to do.
	default:
		noProxyValue = fmt.Sprintf("%s,%s", noProxyValue, ip)
	}
	return noProxyVar, noProxyValue, nil
}

func ValidateGenericDriverFlags(remoteIPAddress, remoteSSHUser, sshKeyToConnectRemote string) {
	if remoteIPAddress == "" || remoteSSHUser == "" || sshKeyToConnectRemote == "" {
		msg := fmt.Sprintf("Generic driver require additional information i.e. IP address of remote machine, path to ssh key and ssh username of the remote host.\n"+
			"Enable experimental features of Minisift and provide following flags to use generic driver:\n"+
			"--%s string\n--%s string\n--%s string\n", configCmd.RemoteIPAddress.Name, configCmd.RemoteSSHUser.Name, configCmd.SSHKeyToConnectRemote.Name)
		atexit.ExitWithMessage(1, fmt.Sprintf("Error: %s", msg))
	}
}

// OcClusterDown stop Openshift cluster using oc binary inside the remote machine
func OcClusterDown(hostVm *host.Host) error {
	sshCommander := provision.GenericSSHCommander{Driver: hostVm.Driver}
	cmd := fmt.Sprintf("%s/oc cluster down", minishiftConstants.OcPathInsideVM)
	_, err := sshCommander.SSHCommand(cmd)
	return err
}
