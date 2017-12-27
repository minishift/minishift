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
	"github.com/golang/glog"
	registrationUtil "github.com/minishift/minishift/cmd/minishift/cmd/registration"
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
	forceProfileDeletion bool
)

func runProfileDelete(cmd *cobra.Command, args []string) {
	validateArgs(args)
	profileName := args[0]

	if !cmdUtil.IsValidProfile(profileName) {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error: '%s' is not a valid profile", profileName))
	}

	profileActions.UpdateProfileConstants(profileName)
	profileBaseDir := constants.Minipath

	if profileName == constants.DefaultProfileName {
		atexit.ExitWithMessage(1, fmt.Sprintf("Default profile '%s' can not be deleted", profileName))
	}

	if !forceProfileDeletion {
		var hasConfirmed bool
		if profileActions.GetActiveProfile() == profileName {
			hasConfirmed = pkgUtil.AskForConfirmation("You are deleting the active profile. It will remove the VM and all related artifacts.")
		} else {
			hasConfirmed = pkgUtil.AskForConfirmation("Will remove the VM and all the related artifacts.")
		}
		if !hasConfirmed {
			atexit.Exit(1)
		}
	}

	api := libmachine.NewClient(profileBaseDir, constants.MakeMiniPath("certs"))
	defer api.Close()

	if cmdUtil.VMExists(api, constants.MachineName) {
		registrationUtil.UnregisterHost(api, false, forceProfileDeletion)
		err := cluster.DeleteHost(api)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error deleting '%s': %v", constants.MakeMiniPath("certs"), err.Error()))
			return
		}

		if glog.V(2) {
			fmt.Println(fmt.Sprintf("Deleted: Minishift VM '%s'", constants.MachineName))
		}
	}

	err := os.RemoveAll(profileBaseDir)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error deleting '%s': %v", profileBaseDir, err.Error()))
	}

	if glog.V(2) {
		fmt.Println(fmt.Sprintf("Deleted: '%s'", profileBaseDir))
	}

	fmt.Println(fmt.Sprintf("Profile '%s' deleted successfully.", profileName))

	// When active profile is deleted, reset the active profile to default profile
	if profileActions.GetActiveProfile() == profileName {
		fmt.Println(fmt.Sprintf("Switching to default profile '%s' as the active profile.", constants.DefaultProfileName))
		err = profileActions.SetDefaultProfileActive()
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
	}
}

func init() {
	profileDeleteCmd.Flags().BoolVarP(&forceProfileDeletion, "force", "f", false, "Forces the deletion of profile and related files in MINISHIFT_HOME.")
	ProfileCmd.AddCommand(profileDeleteCmd)
}
