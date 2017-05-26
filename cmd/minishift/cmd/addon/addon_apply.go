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

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	routingSuffix = "routing-suffix"
)

var addonsApplyCmd = &cobra.Command{
	Use:   "apply ADDON_NAME ...",
	Short: "Executes the specified add-ons.",
	Long:  "Executes the specified add-ons. The command works with both enabled and disabled add-ons.",
	Run:   runApplyAddon,
}

func init() {
	AddonsCmd.AddCommand(addonsApplyCmd)
}

func runApplyAddon(cmd *cobra.Command, args []string) {

	if len(args) == 0 {
		atexit.ExitWithMessage(1, emptyAddOnError)
	}

	addOnManager := GetAddOnManager()
	for i := range args {
		addonName := args[i]
		if !addOnManager.IsInstalled(addonName) {
			atexit.ExitWithMessage(0, fmt.Sprintf(noAddOnMessage, addonName))
		}
	}

	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	util.ExitIfUndefined(api, constants.MachineName)

	host, err := api.Load(constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	util.ExitIfNotRunning(host.Driver, constants.MachineName)

	ip, err := host.Driver.GetIP()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP: %s", err.Error()))
	}

	ocPath := minishiftConfig.InstanceConfig.OcPath
	routingSuffix := viper.GetString(routingSuffix)
	kubeConfigPath := constants.KubeConfigPath
	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}

	for i := range args {
		addonName := args[i]
		addon := addOnManager.Get(addonName)
		addonContext := GetExecutionContext(ip, routingSuffix, ocPath, kubeConfigPath, sshCommander)
		err = addOnManager.ApplyAddOn(addon, addonContext)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprint("Error executing addon commands: ", err))
		}
	}
}
