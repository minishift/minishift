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

	"github.com/minishift/minishift/out/bindata"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

const (
	unspecifiedSourceError   = "You need to specify the source for the addon."
	failedPluginInstallation = "Addon installation failed with error: %s"

	forceFlag    = "force"
	enableFlag   = "enable"
	defaultsFlag = "defaults"
)

var (
	defaultAssets = []string{"anyuid", "cluster-admin"}
	force         bool
	enable        bool
	defaults      bool
)

var addonsInstallCmd = &cobra.Command{
	Use:   "install [SOURCE]",
	Short: "Installs the specified addon",
	Long:  "Verifies and installs the addon given by a file path",
	Run:   runInstallAddon,
}

func init() {
	addonsInstallCmd.Flags().BoolVar(&force, forceFlag, false, "Forces the installation of the addon even if the addon already has been installed before.")
	addonsInstallCmd.Flags().BoolVar(&enable, enableFlag, false, "If set to true installs and enables the specified addon using the default priority.")
	addonsInstallCmd.Flags().BoolVar(&defaults, defaultsFlag, false, "If set to true installs Minishift's default addons.")
	AddonsCmd.AddCommand(addonsInstallCmd)
}

func runInstallAddon(cmd *cobra.Command, args []string) {
	addOnManager := GetAddOnManager()
	if defaults {
		unpackAddons(addOnManager.BaseDir())
		fmt.Println(fmt.Sprintf("Default addons %s installed", strings.Join(defaultAssets, ", ")))
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

	if enable {
		enableAddon(addOnManager, addOnName, 0)
	}
	fmt.Println(fmt.Sprintf("Addon %s installed", addOnName))
}

func unpackAddons(dir string) {
	for _, asset := range defaultAssets {
		err := bindata.RestoreAssets(dir, asset)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Unable to install default addons: %s", err.Error()))
		}
	}
}
