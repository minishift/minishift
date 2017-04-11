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
	"github.com/minishift/minishift/pkg/minikube/constants"
	hostfolderActions "github.com/minishift/minishift/pkg/minishift/hostfolder"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var hostfolderUmountCmd = &cobra.Command{
	Use:   "umount HOSTFOLDER_NAME",
	Short: "Unmount a host folder from the running OpenShift cluster.",
	Long:  `Unmount a host folder from the running OpenShift cluster. This command does not remove the host folder definition or the host folder itself.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			atexit.ExitWithMessage(1, "Usage: minishift hostfolder umount HOSTFOLDER_NAME")
		}

		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		util.ExitIfUndefined(api, constants.MachineName)

		host, err := api.Load(constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		util.ExitIfNotRunning(host.Driver)

		err = hostfolderActions.Umount(host.Driver, args[0])
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error to unmount host folder: %s", err.Error()))
		}
	},
}

func init() {
	HostfolderCmd.AddCommand(hostfolderUmountCmd)
}
