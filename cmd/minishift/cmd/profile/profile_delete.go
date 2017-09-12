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
	"os"

	"github.com/docker/machine/libmachine"
	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	pkgUtil "github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var (
	profileDeleteCmd = &cobra.Command{
		Use:   "delete PROFILE_NAME",
		Short: "delete profiles.",
		Long:  "deletes an existing profile.",
		Run:   runProfileDelete,
	}
	force bool
)

func runProfileDelete(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		atexit.ExitWithMessage(1, emptyProfileMessage)
	} else if len(args) > 1 {
		atexit.ExitWithMessage(1, extraArgumentMessage)
	}

	profileName := args[0]
	var defaultProfile bool

	if !cmdUtil.IsValidProfile(profileName) {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error: '%s' is not a valid profile", profileName))
	}

	profileActions.UpdateProfileConstants(profileName)
	profileBaseDir := constants.GetProfileHomeDir()

	if profileName == constants.DefaultProfileName {
		defaultProfile = true
	}
	if defaultProfile {
		atexit.ExitWithMessage(1, fmt.Sprintf("Default profile '%s' can not be deleted.", profileName))
	}
	if !force {
		hasConfirmed := pkgUtil.AskForConfirmation("Will remove the VM and all the artifacts related to the profile.")
		if !hasConfirmed {
			atexit.Exit(1)
		}
	}

	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	exists := cmdUtil.VMExists(api, constants.MachineName)
	if !exists {
		fmt.Println(fmt.Sprintf("VM for profile '%s' does not exist", profileName))
	} else {
		cmdUtil.UnregisterHost(api, false)
		err := cluster.DeleteHost(api)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error deleting '%s': %v", constants.MakeMiniPath("machines"), err.Error()))
		}
	}

	err := os.RemoveAll(profileBaseDir)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error deleting '%s': %v", profileBaseDir, err.Error()))
	} else {
		fmt.Println("Deleted: ", profileBaseDir)
	}

	//When active profile is deleted, reset the active profile to default profile
	if profileActions.GetActiveProfile() == profileName {
		err = profileActions.SetActiveProfile(constants.DefaultProfileName)
	}
}

func init() {
	profileDeleteCmd.Flags().BoolVar(&force, "force", false, "Forces the deletion of profile specific files in MINISHIFT_HOME.")
	ProfileCmd.AddCommand(profileDeleteCmd)
}
