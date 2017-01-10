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

package cmd

import (
	"fmt"
	"os"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/registration"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes the Minishift VM.",
	Long: `Deletes the Minishift VM, including the local OpenShift cluster and all associated files.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Deleting the Minishift VM...")
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()
		host, err := api.Load(constants.MachineName)
		if err != nil {
			fmt.Println("Error occurred while deleting the VM: ", err)
			os.Exit(1)
		}

		if !drivers.MachineInState(host.Driver, state.Stopped)() {
			// Unregister Host VM
			if err := registration.UnregisterHostVM(host, RegistrationParameters); err != nil {
				fmt.Printf("Error unregistring the VM: %s", err)
				os.Exit(1)
			}
		}

		if err := cluster.DeleteHost(api); err != nil {
			fmt.Println("Error deleting the VM: ", err)
			os.Exit(1)
		}
		fmt.Println("Minishift VM deleted.")
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
