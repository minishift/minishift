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
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes the Minishift VM.",
	Long:  `Deletes the Minishift VM, including the local OpenShift cluster and all associated files.`,
	Run:   runDelete,
}

func runDelete(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	exists, _ := api.Exists(constants.MachineName)
	if !exists {
		fmt.Println("Currently no Minishift VM defined.")
		return
	}

	fmt.Println("Deleting the Minishift VM...")
	if err := cluster.DeleteHost(api); err != nil {
		fmt.Println("Error deleting the VM: ", err)
		atexit.Exit(1)
	}

	if err := instanceState.Config.Delete(); err != nil {
		fmt.Println(fmt.Sprintf("Error deleting %s: ", instanceState.Config.FilePath), err)
		atexit.Exit(1)
	}

	exists = filehelper.Exists(constants.KubeConfigPath)
	if exists {
		if err := os.Remove(constants.KubeConfigPath); err != nil {
			fmt.Println(fmt.Sprintf("Error deleting '%s'", constants.KubeConfigPath))
			atexit.Exit(1)
		}
	}

	fmt.Println("Minishift VM deleted.")
}

func init() {
	viper.BindPFlags(deleteCmd.Flags())
	RootCmd.AddCommand(deleteCmd)
}
