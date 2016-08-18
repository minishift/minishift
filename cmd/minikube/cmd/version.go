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

	"github.com/spf13/cobra"

	"github.com/jimmidyson/minishift/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of minishift.",
	Long:  `Print the version of minishift.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Explicitly disable update checking for the version command
		enableUpdateNotification = false

		RootCmd.PersistentPreRun(cmd, args)
	},
	Run: func(command *cobra.Command, args []string) {
		fmt.Println("minishift version:", version.GetVersion())
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
