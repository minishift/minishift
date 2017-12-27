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

	"strings"

	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

const (
	unspecifiedSourceError   = "You must specify the source of the add-on."
	failedPluginInstallation = "Add-on installation failed with the error: %s"

	forceFlag    = "force"
	enableFlag   = "enable"
	defaultsFlag = "defaults"
)

var (
	force    bool
	enable   bool
	defaults bool
)

var addonsInstallCmd = &cobra.Command{
	Use:   "install [SOURCE]",
	Short: "Installs the specified add-on.",
	Long:  "Installs the add-on from the specified file path and verifies the installation.",
	Run:   runInstallAddon,
}

func init() {
	addonsInstallCmd.Flags().BoolVarP(&force, forceFlag, "f", false, "Forces the installation of the add-on even if the add-on was previously installed.")
	addonsInstallCmd.Flags().BoolVar(&enable, enableFlag, false, "If true, installs and enables the specified add-on with the default priority.")
	addonsInstallCmd.Flags().BoolVar(&defaults, defaultsFlag, false,
		fmt.Sprintf("If true, installs all default add-ons. (%s)",
			strings.Join(util.DefaultAssets, ", ")))
	AddonsCmd.AddCommand(addonsInstallCmd)
}

func runInstallAddon(cmd *cobra.Command, args []string) {
	addOnManager := GetAddOnManager()
	if defaults {
		util.UnpackAddons(addOnManager.BaseDir())
		fmt.Println(fmt.Sprintf("Default add-ons '%s' installed", strings.Join(util.DefaultAssets, ", ")))
		return
	}

	if len(args) != 1 {
		atexit.ExitWithMessage(1, unspecifiedSourceError)
	}

	source := args[0]

	addOnName, err := addOnManager.Install(source, force)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf(failedPluginInstallation, err.Error()))
	}
	fmt.Println(fmt.Sprintf("Addon '%s' installed", addOnName))

	if enable {
		// need to get a new manager
		addOnManager := GetAddOnManager()
		enableAddon(addOnManager, addOnName, 0)
	}
}
