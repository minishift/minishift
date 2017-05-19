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

	"github.com/minishift/minishift/pkg/minishift/update"
	"github.com/minishift/minishift/pkg/util/os/atexit"

	"github.com/minishift/minishift/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update to latest version of Minishift.",
	Long:  `Checks for the latest version of Minishift, prompt the user and update the binary if user answers with 'y'.`,
	Run:   runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) {
	proxyConfig, err := util.NewProxyConfig(viper.GetString(httpProxy), viper.GetString(httpsProxy), "")
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	if proxyConfig.IsEnabled() {
		proxyConfig.ApplyToEnvironment()
	}

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
			err := update.Update(latestVersion)
			if err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Update failed: %s", err))
			}

			fmt.Printf("\nUpdated successfully to minishift v%s.\n", latestVersion)
		}
	} else {
		fmt.Printf("Nothing to update.\nAlready using latest version: %s.\n", latestVersion)
	}
}

func init() {
	RootCmd.AddCommand(updateCmd)
	updateCmd.Flags().AddFlag(httpProxyFlag)
	updateCmd.Flags().AddFlag(httpsProxyFlag)
}
