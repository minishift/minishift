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

	"github.com/golang/glog"
	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minishift/config"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var profileSetCmd = &cobra.Command{
	Use:   "set PROFILE_NAME",
	Short: "Sets the active profile for Minishift.",
	Long:  "Sets the active profile for Minishift. After you set the profile, all commands will use the specified profile by default.",
	Run:   runProfile,
}

func runProfile(cmd *cobra.Command, args []string) {
	validateArgs(args)
	profileName := args[0]

	// Unset the OC CLI context if present for the active profile
	if cmdUtil.DoesVMExist(config.AllInstancesConfig.ActiveProfile) {
		err := cmdUtil.RemoveCurrentContext()
		if err != nil {
			if glog.V(2) {
				fmt.Println(fmt.Sprintf("%s", err.Error()))
			}
		}
	}

	if cmdUtil.IsValidProfile(profileName) {
		err := profileActions.SetActiveProfile(profileName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
		fmt.Printf("Profile '%s' set as active profile.\n", profileName)

		err = cmdUtil.SetOcContext(profileName)
		if err != nil {
			if glog.V(2) {
				fmt.Println(fmt.Sprintf("%s", err.Error()))
				fmt.Println(fmt.Sprintf("oc testing.cli context could not changed for '%s'.", profileName))
			}
		}
	} else {
		err := profileActions.SetActiveProfile(profileName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
		fmt.Printf("Profile '%s' set as active profile.\n", profileName)
	}

}

func init() {
	ProfileCmd.AddCommand(profileSetCmd)
}
