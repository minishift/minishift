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

package cmd

import (
	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

// sshCmd represents the docker-machine ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh [-- COMMAND]",
	Short: "Log in to or run a command on a Minishift VM with SSH.",
	Long:  "Log in to or run a command on a Minishift VM with SSH. This command is similar to 'docker-machine ssh'.",
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()
		err := cluster.CreateSSHShell(api, args)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Cannot establish SSH connection to the VM: %s", err.Error()))
		}
	},
}

func init() {
	RootCmd.AddCommand(sshCmd)
}
