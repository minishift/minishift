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

package hostfolder

import (
	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var umountCmd = &cobra.Command{
	Use:   "umount HOST_FOLDER_NAME",
	Short: "Umount a host folder from the running Minishift VM.",
	Long:  `Umount a host folder from the running Minishift VM. This command does not remove the host folder config or the host folder itself.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			atexit.ExitWithMessage(1, "Usage: minishift hostfolder umount HOST_FOLDER_NAME")
		}

		api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
		defer api.Close()

		util.ExitIfUndefined(api, constants.MachineName)

		host, err := api.Load(constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		util.ExitIfNotRunning(host.Driver, constants.MachineName)

		hostFolderManager := getHostFolderManager()

		name := args[0]
		err = hostFolderManager.Umount(host.Driver, name)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
	},
}

func init() {
	HostFolderCmd.AddCommand(umountCmd)
}
