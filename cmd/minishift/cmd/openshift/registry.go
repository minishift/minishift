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

package openshift

import (
	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/cmd/addon"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/minishift/openshift"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry [flags]",
	Short: "Prints the host name and port number of the OpenShift registry to the standard output.",
	Long:  `Prints the host name and port number of the OpenShift Docker registry in the format: 'host:port'`,
	Run: func(cmd *cobra.Command, args []string) {

		//Check if Minishift VM is running
		api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
		defer api.Close()

		util.ExitIfUndefined(api, constants.MachineName)

		host, err := api.Load(constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		util.ExitIfNotRunning(host.Driver, constants.MachineName)
		registryRoute := addon.GetAddOnManager().Get("registry-route")
		registryAddonEnabled := true
		if registryRoute == nil || !registryRoute.IsEnabled() {
			registryAddonEnabled = false
		}
		openshiftVersion := instanceState.InstanceStateConfig.OpenshiftVersion
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
		registryInfo, err := openshift.GetDockerRegistryInfo(registryAddonEnabled, openshiftVersion)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		fmt.Println(registryInfo)
	},
}

func init() {
	OpenShiftCmd.AddCommand(registryCmd)
}
