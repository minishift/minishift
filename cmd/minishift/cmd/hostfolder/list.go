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
	"github.com/docker/machine/libmachine/drivers"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists the defined host folders.",
	Long:  `Lists an overview of the defined host folders that can be mounted into the Minishift VM.`,
	Run: func(cmd *cobra.Command, args []string) {
		hostFolderManager := getHostFolderManager()

		api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
		defer api.Close()

		mountInfos, err := hostFolderManager.List(getDriver(api))
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		w := tabwriter.NewWriter(os.Stdout, 4, 8, 3, ' ', 0)
		fmt.Fprintln(w, "Name\tType\tSource\tMountpoint\tMounted")

		for _, info := range mountInfos {
			mounted := "N"
			if info.Mounted {
				mounted = "Y"
			}

			fmt.Fprintln(w,
				fmt.Sprintf("%s\t%s\t%s\t%s\t%s",
					info.Name,
					info.Type,
					info.Source,
					info.MountPoint,
					mounted))
		}

		w.Flush()
	},
}

func getDriver(api *libmachine.Client) drivers.Driver {
	if !util.VMExists(api, constants.MachineName) {
		return nil
	}

	host, err := api.Load(constants.MachineName)
	if err != nil {
		return nil
	}

	return host.Driver
}

func init() {
	HostFolderCmd.AddCommand(listCmd)
}
