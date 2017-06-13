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

	"encoding/json"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

type UpdateMarker struct {
	InstallAddon    bool
	PreviousVersion string
}

const (
	addonForceFlag = "force"
)

var (
	addonForce bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update to latest version of Minishift.",
	Long:  `Checks for the latest version of Minishift, prompt the user and update the binary if user answers with 'y'.`,
	Run:   runUpdate,
}

var (
	confirm string
)

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
		if !addonForce {
			fmt.Printf("A newer version of minishift is available.\nDo you want to update from %s to %s now? [y/N]: ", localVersion, latestVersion)
			fmt.Scanln(&confirm)
		}

		if strings.ToLower(confirm) == "y" {
			err := update.Update(latestVersion)
			if err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Update failed: %s", err))
			}

			fmt.Printf("\nUpdated successfully to minishift v%s.\n", latestVersion)

			markerData := UpdateMarker{InstallAddon: false, PreviousVersion: version.GetVersion()}

			fmt.Print("\nDo you want to update default addons? [y/N]: ")
			fmt.Scanln(&confirm)

			if strings.ToLower(confirm) == "y" {
				markerData.InstallAddon = true
				fmt.Println("Default add-ons will be updated when executing the next minishift command.")
			}
			if err := createUpdateMarker(filepath.Join(constants.Minipath, constants.UpdateMarkerFileName), markerData); err != nil {
				atexit.ExitWithMessage(1, "Failed to create update marker file.")
			}

		}

	} else {
		fmt.Printf("Nothing to update.\nAlready using latest version: %s.\n", latestVersion)
	}
}

func init() {
	RootCmd.AddCommand(updateCmd)
	updateCmd.Flags().AddFlag(httpProxyFlag)
	updateCmd.Flags().AddFlag(httpsProxyFlag)
	updateCmd.Flags().BoolVar(&addonForce, addonForceFlag, false, "Force update the addons post binary update. Otherwise, prompt user to update addons.")
}

func createUpdateMarker(markerPath string, data UpdateMarker) error {
	f, err := os.OpenFile(markerPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		return err
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	f.Write(b)
	return nil
}
