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
	"github.com/spf13/viper"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the running local OpenShift cluster.",
	Long: `Stops the running local OpenShift cluster. This command stops the Minishift
VM but does not delete any associated files. To start the cluster again, use the 'minishift start' command.`,
	Run: runStop,
}

func runStop(cmd *cobra.Command, args []string) {
	fmt.Println("Stopping local OpenShift cluster...")
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	if err := cluster.StopHost(api); err != nil {
		fmt.Println("Error stopping cluster: ", err)
		atexit.ExitWithMessage(1, fmt.Sprintf("Error stopping cluster: %s", err.Error()))
	}
	fmt.Println("Cluster stopped.")
}

func init() {
	viper.BindPFlags(stopCmd.Flags())
	RootCmd.AddCommand(stopCmd)
}
