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

	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

const (
	emptyDisableError       = "An addon name needs to be specified. Use `minishift addons list` to view installed addons."
	noAddOnToDisableMessage = "No addon with name %s installed"
)

var addonsDisableCmd = &cobra.Command{
	Use:   "disable ADDON_NAME",
	Short: "Disables the specified addon.",
	Long:  "Disables the specified addon from being run after cluster creation.",
	Run:   runDisableAddon,
}

func init() {
	AddonsCmd.AddCommand(addonsDisableCmd)
}

func runDisableAddon(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		atexit.ExitWithMessage(1, emptyDisableError)
	}

	addonName := args[0]
	addOnManager := GetAddOnManager()

	if !addOnManager.IsInstalled(addonName) {
		fmt.Println(fmt.Sprintf(noAddOnToDisableMessage, addonName))
		return
	}

	addOnConfig, err := addOnManager.Disable(addonName)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Unable to disable plugin %s: %s", addonName, err.Error()))
	}

	addOnConfigMap := getAddOnConfiguration()
	addOnConfigMap[addOnConfig.Name] = addOnConfig
	writeAddOnConfig(addOnConfigMap)
}
