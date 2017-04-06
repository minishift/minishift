/*
Copyright (C) 2016 Red Hat, Inc.

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
	"github.com/spf13/cobra"

	"github.com/docker/machine/libmachine"

	"fmt"

	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/minishift/minishift/pkg/minishift/openshift"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

const (
	configTargetFlag         = "target"
	unknownConfigTargetError = "Unkown config target. Only 'master' and 'node' are supported."
)

var (
	configTarget string
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Displays the specified OpenShift configuration resource.",
	Long:  "Displays the specified OpenShift configuration resource.",
	Run:   runViewConfig,
}

func init() {
	viewCmd.Flags().StringVar(&configTarget, configTargetFlag, "master", "Target configuration to display. Either 'master' or 'node'.")
	configCmd.AddCommand(viewCmd)
}

func runViewConfig(cmd *cobra.Command, args []string) {
	configFileTarget := determineTarget(configTarget)
	if configFileTarget == openshift.UNKNOWN {
		atexit.ExitWithMessage(1, unknownConfigTargetError)
	}

	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		atexit.ExitWithMessage(1, nonExistentMachineError)
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		atexit.ExitWithMessage(1, unableToRetrieveIpError)
	}
	configFileTarget.SetIp(ip)

	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}
	dockerCommander := docker.NewVmDockerCommander(sshCommander)

	out, err := openshift.ViewConfig(configFileTarget, dockerCommander)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Unable to display OpenShift configuration: %s", err.Error()))
	}

	fmt.Println(out)
}
