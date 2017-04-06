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
	"github.com/minishift/minishift/pkg/minikube/constants"

	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

// ipCmd represents the ip command
var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "Gets the IP address of the running cluster.",
	Long:  `Gets the IP address of the running cluster and prints it to standard output.`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()
		host, err := api.Load(constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP: %s", err.Error()))
		}
		ip, err := host.Driver.GetIP()
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP: %s", err.Error()))
		}
		fmt.Println(ip)
	},
}

func init() {
	RootCmd.AddCommand(ipCmd)
}
