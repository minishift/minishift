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
	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	pkgUtil "github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var (
	// deleteCmd represents the delete command
	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Deletes the Minishift VM.",
		Long:  `Deletes the Minishift VM, including the local OpenShift cluster and all associated files.`,
		Run:   runDelete,
	}

	forceMachineDeletion bool
	clearCache           bool
)

func runDelete(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	util.ExitIfUndefined(api, constants.MachineName)

	if !forceMachineDeletion {
		hasConfirmed := pkgUtil.AskForConfirmation(fmt.Sprintf("You are deleting the Minishift VM: '%s'.", constants.MachineName))
		if !hasConfirmed {
			atexit.Exit(1)
		}
	}

	if clearCache {
		cachePath := constants.MakeMiniPath("cache")
		err := os.RemoveAll(cachePath)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error deleting Minishift cache: %v", err))
		} else {
			fmt.Printf("Removed the cache at: %s\n", cachePath)
		}
	}

	// Unregistration, do not allow to be skipped
	cmdUtil.UnregisterHost(api, false)

	fmt.Println("Deleting the Minishift VM...")
	if err := cluster.DeleteHost(api); err != nil {
		handleFailedHostDeletion(err)
	}

	removeInstanceConfigs()
	fmt.Println("Minishift VM deleted.")
}

func handleFailedHostDeletion(err error) {

	if forceMachineDeletion {
		err := os.RemoveAll(constants.MakeMiniPath("machines"))
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error deleting '%s': %v", constants.MakeMiniPath("machines"), err))
		}
	} else {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error deleting the Minishift VM: %v", err))
	}
}

func removeInstanceConfigs() {
	exists := filehelper.Exists(minishiftConfig.InstanceConfig.FilePath)
	if exists {
		if err := minishiftConfig.InstanceConfig.Delete(); err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintln(fmt.Sprintf("Error deleting %s: ", minishiftConfig.InstanceConfig.FilePath), err))
		}
	}
	exists = filehelper.Exists(constants.KubeConfigPath)
	if exists {
		if err := os.Remove(constants.KubeConfigPath); err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error deleting '%s'", constants.KubeConfigPath))
		}
	}
}

func init() {
	deleteCmd.Flags().BoolVar(&forceMachineDeletion, "force", false, "Forces the deletion of the VM specific files in MINISHIFT_HOME.")
	deleteCmd.Flags().BoolVar(&clearCache, "clear-cache", false, "Deletes all cached artifacts as part of the VM deletion.")
	RootCmd.AddCommand(deleteCmd)
}
