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
	emptyDisableError       = "You must specify an add-on name. Use `minishift addons list` to view installed add-ons."
	noAddOnToDisableMessage = "No add-on with the name %s is installed."
)

var addonsDisableCmd = &cobra.Command{
	Use:   "disable ADDON_NAME",
	Short: "Disables the specified add-on.",
	Long:  "Disables the specified add-on and prevents applying the add-on the next time a cluster is created.",
	Run:   runDisableAddon,
}

func init() {
	AddonsCmd.AddCommand(addonsDisableCmd)
}

func runDisableAddon(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		atexit.ExitWithMessage(1, emptyDisableError)
	}

	addOnName := args[0]
	addOnManager := GetAddOnManager()

	if !addOnManager.IsInstalled(addOnName) {
		atexit.ExitWithMessage(0, fmt.Sprintf(noAddOnToDisableMessage, addOnName))
	}

	addOnConfig, err := addOnManager.Disable(addOnName)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Unable to disable add-on %s: %s", addOnName, err.Error()))
	} else {
		fmt.Println(fmt.Sprintf("Addon '%s' disabled", addOnName))
	}

	addOnConfigMap := getAddOnConfiguration()
	addOnConfigMap[addOnConfig.Name] = addOnConfig
	writeAddOnConfig(addOnConfigMap)
}
