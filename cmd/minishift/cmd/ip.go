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
	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftNetwork "github.com/minishift/minishift/pkg/minishift/network"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var (
	configureAsStatic  bool
	configureAsDynamic bool
)

// ipCmd represents the ip command
var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "Gets the IP address of the running cluster.",
	Long:  `Gets the IP address of the running cluster and prints it to standard output.`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
		defer api.Close()

		cmdUtil.ExitIfUndefined(api, constants.MachineName)

		if configureAsStatic && configureAsDynamic {
			atexit.ExitWithMessage(1, "Invalid options specified")
		}

		host, err := api.Load(constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP: %s", err.Error()))
		}
		cmdUtil.ExitIfNotRunning(host.Driver, constants.MachineName)

		if configureAsDynamic {
			minishiftNetwork.ConfigureDynamicAssignment(host.Driver)
		} else if configureAsStatic {
			minishiftNetwork.ConfigureStaticAssignment(host.Driver)
		} else {
			ip, err := minishiftNetwork.GetIP(host.Driver)
			if err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP: %s", err.Error()))
			}
			fmt.Println(ip)
		}
	},
}

func init() {
	ipCmd.Flags().BoolVar(&configureAsStatic, "set-static", false, "Sets the current assigned IP address as static address for the instance")
	ipCmd.Flags().BoolVar(&configureAsDynamic, "set-dhcp", false, "Sets network configuration to use DHCP to assign IP address to the instance")

	RootCmd.AddCommand(ipCmd)
}
