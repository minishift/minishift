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

package addon

import (
	"fmt"

	"github.com/minishift/minishift/pkg/minishift/addon/manager"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var addonsUnInstallCmd = &cobra.Command{
	Use:   "uninstall ADDON_NAME",
	Short: "Uninstalls the specified add-on.",
	Long:  "Uninstalls the specified add-on.",
	Run:   runUnInstallAddon,
}

func init() {
	AddonsCmd.AddCommand(addonsUnInstallCmd)
}

func runUnInstallAddon(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		atexit.ExitWithMessage(1, emptyAddOnError)
	}

	addonName := args[0]
	addOnManager := GetAddOnManager()

	if !addOnManager.IsInstalled(addonName) {
		atexit.ExitWithMessage(0, fmt.Sprintf(noAddOnMessage, addonName))
	}

	unInstallAddon(addOnManager, addonName)
}

func unInstallAddon(addOnManager *manager.AddOnManager, addOnName string) {
	err := addOnManager.UnInstall(addOnName)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Cannot Uninstall the add-on %s: %s", addOnName, err.Error()))
	} else {
		fmt.Println(fmt.Sprintf("Add-on '%s' uninstalled", addOnName))
	}
	RemoveAddOnFromConfig(addOnName)
}
