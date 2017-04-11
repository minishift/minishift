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

	"os"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util/filehelper"
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
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	util.ExitIfUndefined(api, constants.MachineName)

	fmt.Println("Deleting the Minishift VM...")
	if err := cluster.DeleteHost(api); err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error deleting the VM: %s", err.Error()))
	}

	if err := minishiftConfig.InstanceConfig.Delete(); err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintln(fmt.Sprintf("Error deleting %s: ", minishiftConfig.InstanceConfig.FilePath), err))
	}

	exists := filehelper.Exists(constants.KubeConfigPath)
	if exists {
		if err := os.Remove(constants.KubeConfigPath); err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error deleting '%s'", constants.KubeConfigPath))
		}
	}

	fmt.Println("Minishift VM deleted.")
}

func init() {
	viper.BindPFlags(deleteCmd.Flags())
	RootCmd.AddCommand(deleteCmd)
}
