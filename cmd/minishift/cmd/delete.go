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

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes the Minishift VM.",
	Long:  `Deletes the Minishift VM, including the local OpenShift cluster and all associated files.`,
	Run:   runDelete,
}

func runDelete(cmd *cobra.Command, args []string) {
	fmt.Println("Deleting the Minishift VM...")
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	setSubcriptionManagerParameters()

	if err := cluster.DeleteHost(api); err != nil {
		fmt.Println("Error deleting the VM: ", err)
		atexit.Exit(1)
	}
	fmt.Println("Minishift VM deleted.")
}

func init() {
	deleteCmd.Flags().String(username, "", "Username for the virtual machine unregistration.")
	deleteCmd.Flags().String(password, "", "Password for the virtual machine unregistration.")

	deleteCmd.Flags().AddFlagSet(subscriptionManagerFlagSet)

	viper.BindPFlags(deleteCmd.Flags())
	RootCmd.AddCommand(deleteCmd)
}
