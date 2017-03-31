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
	"github.com/docker/machine/libmachine/provision"
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
	"os"
)

// version command represent current running openshift version and available one.
var versionCmd = &cobra.Command{
	Use:   "version [command] [flags]",
	Short: "Print the current running openshift version to stdout",
	Long:  `Print the current running openshift version to stdout`,
	Run:   runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		fmt.Println(nonExistentMachineError)
		atexit.Exit(1)
	}
	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}
	dockerCommander := docker.NewVmDockerCommander(sshCommander)
	version, err := dockerCommander.Exec(" ", "origin", "openshift", "version")
	if err != nil {
		glog.Errorln("Error restarting OpenShift cluster: ", err)
		atexit.Exit(1)
	}
	fmt.Fprintln(os.Stdout, version)
}

func init() {
	OpenShiftConfigCmd.AddCommand(versionCmd)
}
