/*
Copyright (C) 2017 Red Hat, Inc.

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

package hostfolder

import (
	"github.com/spf13/cobra"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
)

func isHostRunning(driver drivers.Driver) bool {
	return drivers.MachineInState(driver, state.Running)()
}

var HostfolderCmd = &cobra.Command{
	Use:   "hostfolder SUBCOMMAND [flags]",
	Short: "Manage and control host folders for use by the OpenShift cluster.",
	Long:  `Manage and control host folders for use by the OpenShift cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
