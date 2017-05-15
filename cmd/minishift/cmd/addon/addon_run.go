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
	"path/filepath"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/cache"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/minishift/minishift/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	routingSuffix = "routing-suffix"
)

var addonsRunCmd = &cobra.Command{
	Use:   "run ADDON_NAME",
	Short: "Executes the specified add-on.",
	Long:  "Executes the specified add-on",
	Run:   runAddon,
}

func init() {
	AddonsCmd.AddCommand(addonsRunCmd)
}

func runAddon(cmd *cobra.Command, args []string) {

	//Check if Minishift is running
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	util.ExitIfUndefined(api, constants.MachineName)

	host, err := api.Load(constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	util.ExitIfNotRunning(host.Driver)

	ip, err := host.Driver.GetIP()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP: %s", err.Error()))
	}

	oc := cache.Oc{
		OpenShiftVersion:  version.GetOpenShiftVersion(),
		MinishiftCacheDir: filepath.Join(constants.Minipath, "cache"),
	}
	ocPath := filepath.Join(oc.GetCacheFilepath(), constants.OC_BINARY_NAME)
	routingSuffix := viper.GetString(routingSuffix)
	kubeConfigPath := constants.KubeConfigPath
	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}

	addOnManager := GetAddOnManager()
	err = addOnManager.Apply(GetExecutionContext(ip, routingSuffix, ocPath, kubeConfigPath, sshCommander))
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprint("Error executing addon commands: ", err))
	}
}
