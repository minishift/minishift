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

package profile

import (
	"fmt"
	"path/filepath"

	"github.com/docker/machine/libmachine"
	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	cmdutil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/clusterup"
	"github.com/minishift/minishift/pkg/minishift/oc"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var profileSetCmd = &cobra.Command{
	Use:   "set PROFILE_NAME",
	Short: "Sets the active profile for Minishift.",
	Long:  "Sets the active profile for Minishift. After you set the profile, all commands will use the profile by default",
	Run: func(cmd *cobra.Command, args []string) {
		var doesProfileExist = false
		if len(args) == 0 {
			atexit.ExitWithMessage(1, emptyProfileError)
		} else if len(args) > 1 {
			atexit.ExitWithMessage(1, extraArgumentError)
		}

		profileName := args[0]

		//check if the profile is present in the AllInstancesConfig
		profileList := profileActions.GetProfileNameList()
		for i := range profileList {
			if profileList[i] == profileName {
				doesProfileExist = true
			}
		}
		if !doesProfileExist {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error: '%s' is not a valid profile", profileName))
		}

		setOcCliContext(profileName)
		err := profileActions.SetActiveProfile(profileName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		} else {
			fmt.Println(fmt.Sprintf("Profile set to '%s' successfully", profileName))
		}

	},
}

func setOcCliContext(profileName string) {
	const (
		defaultProject = "myproject"
		defaultUser    = "developer"
	) //This needs to be fixed

	//We need to reassign ProfileName, MachineName, KubeConfigPath as it would have values for the previous
	// profile
	constants.ProfileName = profileName
	constants.MachineName = constants.ProfileName
	constants.Minipath = constants.GetProfileHomeDir()

	constants.KubeConfigPath = filepath.Join(constants.Minipath, "machines", constants.MachineName+"_kubeconfig")

	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()
	util.ExitIfUndefined(api, constants.MachineName)

	host, err := api.Load(constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting information for VM: %s", profileName))
	}

	util.ExitIfNotRunning(host.Driver, constants.MachineName)

	ip, err := host.Driver.GetIP()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting the IP address: %s", err.Error()))
	}

	requestedOpenShiftVersion := viper.GetString(configCmd.OpenshiftVersion.Name)
	ocPath := cmdutil.CacheOc(clusterup.DetermineOcVersion(requestedOpenShiftVersion))

	ocRunner, err := oc.NewOcRunner(ocPath, constants.KubeConfigPath)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error during setting '%s' as active profile: %s", profileName, err.Error()))
	}
	err = ocRunner.AddCliContext(constants.MachineName, ip, defaultUser, defaultProject)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error during setting '%s' as active profile: %s", profileName, err.Error()))
	}

}

func init() {
	ProfileCmd.AddCommand(profileSetCmd)
}
