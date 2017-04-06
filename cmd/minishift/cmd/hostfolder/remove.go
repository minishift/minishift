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
	hostfolderActions "github.com/minishift/minishift/pkg/minishift/hostfolder"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var hostfolderRemoveCmd = &cobra.Command{
	Use:   "remove HOSTFOLDER_NAME",
	Short: "Removes the specified host folder definition.",
	Long:  `Removes the specified host folder definition. This command does not remove the host folder or any data.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			atexit.ExitWithMessage(1, "Usage: minishift hostfolder remove HOSTFOLDER_NAME")
		}

		err := hostfolderActions.Remove(args[0])
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
	},
}

func init() {
	HostfolderCmd.AddCommand(hostfolderRemoveCmd)
}
