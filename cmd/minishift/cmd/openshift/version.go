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
	"os"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	openshiftVersions "github.com/minishift/minishift/pkg/minishift/openshift/version"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

// version command represent current running openshift version and available one.
var versionCmd = &cobra.Command{
	Use:   "version [command] [flags]",
	Short: "Prints the current running OpenShift version to the standard output.",
	Long:  `Prints the current running OpenShift version to the standard output.`,
	Run:   runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer api.Close()

	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		atexit.ExitWithMessage(1, nonExistentMachineError)
	}
	version, err := openshiftVersions.GetOpenshiftVersion(host)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting the OpenShift cluster version: %s", err.Error()))
	}
	fmt.Fprint(os.Stdout, version)
}

func init() {
	OpenShiftCmd.AddCommand(versionCmd)
}
