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

	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var profileSetCmd = &cobra.Command{
	Use:   "set PROFILE_NAME",
	Short: "Sets the active profile for Minishift.",
	Long:  "Sets the active profile for Minishift. After you set the profile, all commands will use the profile by default",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			atexit.ExitWithMessage(1, emptyProfileError)
		} else if len(args) > 1 {
			atexit.ExitWithMessage(1, extraArgumentError)
		}

		profileName := args[0]

		if !cmdUtil.IsValidProfile(profileName) {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error: '%s' is not a valid profile", profileName))
		}

		err := profileActions.SetActiveProfile(profileName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		} else {
			fmt.Println(fmt.Sprintf("Profile '%s' set as active profile", profileName))
		}
		err = cmdUtil.SetOcContext(profileName)
		if err != nil {
			fmt.Println(fmt.Sprintf("oc cli context could not set to '%s': %s", profileName, err.Error()))
		}
	},
}

func init() {
	ProfileCmd.AddCommand(profileSetCmd)
}
