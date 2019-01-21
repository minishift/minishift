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
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minishift/clusterup"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/os/atexit"

	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	cmdState "github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/out/bindata"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	openshiftVersion "github.com/minishift/minishift/pkg/minishift/openshift/version"
	"github.com/minishift/minishift/pkg/util/shell"
	"github.com/minishift/minishift/pkg/version"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	DefaultAssets = []string{
		"anyuid",
		"admin-user",
		"xpaas",
		"registry-route",
		"che",
		"htpasswd-identity-provider",
		"admissions-webhook",
		"redhat-registry-login",
	}
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
	profileDirs := cmdState.GetMinishiftDirsStructure(constants.GetProfileHomeDir(profileName))

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
			"Provide following flags to use generic driver:\n"+
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

func GetOpenShiftReleaseVersion() (string, error) {
	tag := viper.GetString(configCmd.OpenshiftVersion.Name)
	// tag is in the form of vMajor.minor.patch e.g v3.9.0
	if tag == fmt.Sprintf("%slatest", constants.VersionPrefix) {
		tags, err := openshiftVersion.GetGithubReleases()
		if err != nil {
			return "", err
		}

		sortedtags, err := openshiftVersion.OpenShiftTagsByAscending(tags, constants.MinimumSupportedOpenShiftVersion, version.GetOpenShiftVersion())
		if err != nil {
			return "", err
		}

		return sortedtags[len(sortedtags)-1], nil
	}

	return tag, nil
}

// determineClusterUpParameters returns a map of flag names and values for the cluster up call.
func DetermineClusterUpParameters(config *clusterup.ClusterUpConfig, DockerbridgeSubnet string, clusterUpFlagSet *flag.FlagSet) map[string]string {
	clusterUpParams := make(map[string]string)
	// Set default value for base config for 3.10
	clusterUpParams["base-dir"] = minishiftConstants.BaseDirInsideInstance
	if viper.GetString(configCmd.ImageName.Name) == "" {
		imagetag := fmt.Sprintf("'%s:%s'", minishiftConstants.ImageNameForClusterUpImageFlag, config.OpenShiftVersion)
		viper.Set(configCmd.ImageName.Name, imagetag)
	}
	// Add docker bridge subnet to no-proxy before passing to oc cluster up
	if viper.GetString(configCmd.NoProxyList.Name) != "" {
		viper.Set(configCmd.NoProxyList.Name, fmt.Sprintf("%s,%s", DockerbridgeSubnet, viper.GetString(configCmd.NoProxyList.Name)))
	}

	viper.Set(configCmd.RoutingSuffix.Name, config.RoutingSuffix)
	viper.Set(configCmd.PublicHostname.Name, config.PublicHostname)
	clusterUpFlagSet.VisitAll(func(flag *flag.Flag) {
		if viper.IsSet(flag.Name) {
			value := viper.GetString(flag.Name)
			key := flag.Name
			clusterUpParams[key] = value
		}
	})

	return clusterUpParams
}

func EnsureConfigFileExists(configPath string) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		jsonRoot := []byte("{}")
		f, err := os.Create(configPath)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Cannot create file '%s': %s", configPath, err))
		}
		defer f.Close()
		_, err = f.Write(jsonRoot)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Cannot encode config '%s': %s", configPath, err))
		}
	}
}

func CreateMinishiftDirs(dirs *cmdState.MinishiftDirs) {
	dirPaths := reflect.ValueOf(*dirs)

	for i := 0; i < dirPaths.NumField(); i++ {
		path := dirPaths.Field(i).Interface().(string)
		if err := os.MkdirAll(path, 0777); err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error creating directory: %s", path))
		}
	}
}
