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
	"github.com/blang/semver"
	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	cmdutil "github.com/minishift/minishift/cmd/minishift/cmd/util"
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
	updateForceFlag = "force"
	addonForceFlag  = "update-addons"
)

var (
	addonForce  bool
	force       bool
	versionFlag string
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates Minishift to the latest version.",
	Long:  `Checks for the latest version of Minishift, prompts the user, and updates the binary if the user answers 'y'.`,
	Run:   runUpdate,
}

var (
	addonForceConfirm string
	forceConfirm      string
	addonConfirm      string
)

func runUpdate(cmd *cobra.Command, args []string) {
	proxyConfig, err := util.NewProxyConfig(viper.GetString(configCmd.HttpProxy.Name), viper.GetString(configCmd.HttpsProxy.Name), "")
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	if proxyConfig.IsEnabled() {
		proxyConfig.ApplyToEnvironment()
	}

	currentVersion, err := update.CurrentVersion()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Update failed: %s", err))
	}
	versionToUpdate, err := update.LatestVersion()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Update failed: %s", err))
	}
	if versionFlag != "" {
		versionToUpdate, err = semver.Make(versionFlag)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Update failed: %s", err))
		}
	}

	if versionToUpdate.Major > currentVersion.Major {
		fmt.Println("The latest version is not compatible with the current version. Follow the uninstallation procedure at https://docs.okd.io/latest/minishift/getting-started/uninstalling.html#uninstall-instructions.")
	}

	performUpdate(currentVersion, versionToUpdate)
}

func init() {
	RootCmd.AddCommand(updateCmd)
	updateCmd.Flags().AddFlag(cmdutil.HttpProxyFlag)
	updateCmd.Flags().AddFlag(cmdutil.HttpsProxyFlag)
	updateCmd.Flags().BoolVarP(&force, updateForceFlag, "f", false, "Force update the binary.")
	updateCmd.Flags().BoolVarP(&addonForce, addonForceFlag, "", false, "Force update the add-ons after the binary update. Otherwise, prompt the user to update add-ons.")
	updateCmd.Flags().StringVar(&versionFlag, "version", "", "Specify the version to update (without 'v')")
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

func performUpdate(currentVersion, versionToUpdate semver.Version) {
	if update.IsNewerVersion(currentVersion, versionToUpdate) {
		if !force {
			fmt.Printf("Do you want to update from %s to %s now? [y/N]: ", currentVersion, versionToUpdate)
			fmt.Scanln(&forceConfirm)

			if strings.ToLower(forceConfirm) == "y" {
				updateToVersion(versionToUpdate)
			}
		} else {
			updateToVersion(versionToUpdate)
		}
	} else {
		fmt.Printf("Nothing to update.\nAlready using the latest version: %s.\n", versionToUpdate)
	}
}

func performAddonUpdate(versionToUpdate semver.Version) {
	markerData := UpdateMarker{InstallAddon: false, PreviousVersion: version.GetMinishiftVersion()}
	addonLocationForRelease := fmt.Sprintf("https://github.com/minishift/minishift/tree/v%s/addons", versionToUpdate)

	if addonForce {
		updateAddon(markerData)
	} else {
		fmt.Printf("\nCurrent Installed add-ons are locally present at: %s\n", filepath.Join(constants.Minipath, "addons"))
		fmt.Printf("The add-ons for %s available at: %s\n", versionToUpdate, addonLocationForRelease)
		fmt.Printf("\nDo you want to update the default add-ons? [y/N]: ")
		fmt.Scanln(&addonConfirm)

		if strings.ToLower(addonConfirm) == "y" {
			updateAddon(markerData)
		}
	}
}

func updateToVersion(versionToUpdate semver.Version) {
	if err := update.Update(versionToUpdate); err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Update failed: %s", err))
	}
	fmt.Printf("\nUpdated successfully to Minishift version %s.\n", versionToUpdate)

	performAddonUpdate(versionToUpdate)
}

func updateAddon(marker UpdateMarker) {
	marker.InstallAddon = true
	fmt.Println("Default add-ons will be updated the next time you run any 'minishift' command.")
	if err := createUpdateMarker(filepath.Join(constants.Minipath, constants.UpdateMarkerFileName), marker); err != nil {
		atexit.ExitWithMessage(1, "Failed to create update marker file.")
	}
}
