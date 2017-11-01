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
	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var (
	mountAll bool
)

var mountCmd = &cobra.Command{
	Use:   "mount HOST_FOLDER_NAME",
	Short: "Mounts the specified host folder into the Minishift VM.",
	Long:  `Mounts the specified host folder into the Minishift VM. You can set the 'all' flag to mount all of the defined host folders.`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
		defer api.Close()

		util.ExitIfUndefined(api, constants.MachineName)

		host, err := api.Load(constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		util.ExitIfNotRunning(host.Driver, constants.MachineName)

		hostFolderManager := getHostFolderManager()
		err = nil
		if mountAll {
			fmt.Println("-- Mounting host folders")
			err = hostFolderManager.MountAll(host.Driver)
		} else {
			if len(args) < 1 {
				atexit.ExitWithMessage(1, "Usage: minishift hostfolder mount [HOST_FOLDER_NAME|--all]")
			}
			err = hostFolderManager.Mount(host.Driver, args[0])
		}

		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

	},
}

func init() {
	HostFolderCmd.AddCommand(mountCmd)
	mountCmd.Flags().BoolVarP(&mountAll, "all", "a", false, "Mounts all defined host folders into the Minishift VM.")
}
