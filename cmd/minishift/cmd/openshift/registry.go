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
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/openshift"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry [flags]",
	Short: "Prints host and port of the OpenShift registry.",
	Long:  `Prints host and port of the OpenShift docker registry in the format 'host:port'.`,
	Run: func(cmd *cobra.Command, args []string) {

		//Check if Minishift VM is running
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		util.ExitIfUndefined(api, constants.MachineName)

		host, err := api.Load(constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		util.ExitIfNotRunning(host.Driver)

		registryInfo, err := openshift.GetDockerRegistryInfo()
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		fmt.Println(registryInfo)
	},
}

func init() {
	OpenShiftCmd.AddCommand(registryCmd)
}
