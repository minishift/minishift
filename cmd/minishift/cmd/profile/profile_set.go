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

	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

const (
	emptyProfileError  = "You must provide the profile name. Run `minishift profile list` to view profiles"
	extraArgumentError = "You have provided more arguments than required. You must provide a single profile name"
)

var profileSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets the default profile for Minishift",
	Long:  "Sets the default profile for Minishift. After you set the profile, all commands will use the profile by default",
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

		//Check if the requested profile is already active
		if profileName == profileActions.GetActiveProfile() {
			atexit.ExitWithMessage(1, fmt.Sprintf("'%s' is already set as active profile", profileName))
		}

		err := profileActions.SetActiveProfile(profileName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		} else {
			fmt.Println(fmt.Sprintf("Profile set to '%s' successfully", profileName))
		}
	},
}

func init() {
	ProfileCmd.AddCommand(profileSetCmd)
}
