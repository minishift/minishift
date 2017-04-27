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

package openshift

import (
	"github.com/spf13/cobra"
)

const (
	unableToRetrieveIpError = "Unable to retrieve virtual machine IP."
	nonExistentMachineError = "There is no running OpenShift cluster."
)

var configCmd = &cobra.Command{
	Use:   "config SUBCOMMAND [flags]",
	Short: "Displays or patches OpenShift configuration.",
	Long:  "Displays or patches OpenShift master or node configuration. Patches are supplied in JSON format.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	OpenShiftCmd.AddCommand(configCmd)
}
