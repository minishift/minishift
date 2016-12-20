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
	"log"
	"os"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Gets the logs of the running Minishift instance.",
	Long:  `Gets the logs of the running Minishift instance and prints them to the standard output. You use these log messages to debug Minishift.
	The logs do not contain information about your application code.`,
	//NEEDINFO: Is it stdout? What if I want to output to a file? How can users debug their apps? Anything else? What about log levels (i.e. debug, warning, info)?
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()
		s, err := cluster.GetHostLogs(api)
		if err != nil {
			log.Println("Error getting logs:", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, s)
	},
}

func init() {
	RootCmd.AddCommand(logsCmd)
}
