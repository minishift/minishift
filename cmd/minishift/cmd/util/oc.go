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
	"errors"
	"fmt"
	"path/filepath"

	"github.com/docker/machine/libmachine"
	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/cache"
	"github.com/minishift/minishift/pkg/minishift/clusterup"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/minishift/oc"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/viper"
)

// cacheOc ensures that the oc binary matching the requested OpenShift version is cached on the host
func CacheOc(openShiftVersion string) string {
	ocBinary := cache.Oc{
		OpenShiftVersion:  openShiftVersion,
		MinishiftCacheDir: filepath.Join(constants.Minipath, "cache"),
	}
	if err := ocBinary.EnsureIsCached(); err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error starting the cluster: %v", err))
	}

	// Update MACHINE_NAME.json for oc path
	minishiftConfig.InstanceConfig.OcPath = filepath.Join(ocBinary.GetCacheFilepath(), constants.OC_BINARY_NAME)
	if err := minishiftConfig.InstanceConfig.Write(); err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error updating oc path in config of VM: %v", err))
	}

	return minishiftConfig.InstanceConfig.OcPath
}

func SetOcContext(profileName string) error {
	const (
		defaultProject = "myproject"
		defaultUser    = "developer"
	) //This needs to be fixed

	profileActions.UpdateProfileConstants(profileName)

	// Need create the kube config path for the profile for ocrunner to use it.
	kubeConfigPath := filepath.Join(constants.Minipath, "machines", constants.MachineName+"_kubeconfig")

	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	host, err := api.Load(profileName)
	if err != nil {
		return errors.New(fmt.Sprintf("Error getting information for the VM: %s", profileName))
	}

	running := IsHostRunning(host.Driver)
	if !running {
		return errors.New(fmt.Sprintf("Profile '%s' VM is not running", profileName))
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return errors.New(fmt.Sprintf("Error getting the IP address: '%s'", err.Error()))
	}

	requestedOpenShiftVersion := viper.GetString(configCmd.OpenshiftVersion.Name)
	ocPath := CacheOc(clusterup.DetermineOcVersion(requestedOpenShiftVersion))

	ocRunner, err := oc.NewOcRunner(ocPath, kubeConfigPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error during setting '%s' as active profile: %s", profileName, err.Error()))
	}
	err = ocRunner.AddCliContext(constants.MachineName, ip, defaultUser, defaultProject)
	if err != nil {
		return errors.New(fmt.Sprintf("Error during setting '%s' as active profile: %s", profileName, err.Error()))
	}

	return nil
}
