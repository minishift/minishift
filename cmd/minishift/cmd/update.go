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

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/minishift/minishift/pkg/minishift/update"
	"github.com/minishift/minishift/pkg/util/os/atexit"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Checks for updates in Minishift",
	Long:  `Checks for updates in Minishift and prompts the user to download newer version`,
	Run:   runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) {

	localVersion, err := update.CurrentVersion()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Update failed: %s", err))
	}

	latestVersion, err := update.LatestVersion()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Update failed: %s", err))
	}

	if update.IsNewerVersion(localVersion, latestVersion) {
		fmt.Printf("A newer version of minishift is available.\nDo you want to update from %s to %s now? [y/N]: ", localVersion, latestVersion)

		var confirm string
		fmt.Scanln(&confirm)

		if strings.ToLower(confirm) == "y" {

			err := update.Update(os.Stdout, latestVersion)
			if err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Update failed: %s", err))
			}
		}
	} else {
		fmt.Printf("Nothing to update\nAlready using latest version: %s\n", latestVersion)
	}
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
