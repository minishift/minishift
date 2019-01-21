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

	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	cmdState "github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/config"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var (
	profileCopyCmd = &cobra.Command{
		Use:   "copy SRC_PROFILE_NAME NEW_PROFILE_NAME",
		Short: "Copy configurations from an existing profile to a new profile.",
		Long:  "Copy add-on and config configurations from an existing profile to a new profile.",
		Run:   copyProfile,
	}
)

func copyProfile(cmd *cobra.Command, args []string) {
	validateCopyProfileCmd(args)
	srcProfile := args[0]
	newProfile := args[1]

	if !cmdUtil.IsValidProfile(srcProfile) {
		atexit.ExitWithMessage(1, fmt.Sprintf("Profile %s does not exist", srcProfile))
	}

	if cmdUtil.IsValidProfile(newProfile) {
		atexit.ExitWithMessage(1, fmt.Sprintf("Profile '%s' already exists. You must provide a non-existant profile name", newProfile))
	}
	copySrcToNewProfile(srcProfile, newProfile)

}

func copySrcToNewProfile(srcProfile, newProfile string) {
	srcConfFile := constants.GetProfileConfigFile(srcProfile)
	sourceConfig, err := config.ReadViperConfig(srcConfFile)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	newInstanceDirs := state.GetMinishiftDirsStructure(constants.GetProfileHomeDir(newProfile))
	cmdUtil.CreateMinishiftDirs(newInstanceDirs)
	newConfFile := constants.GetProfileConfigFile(newProfile)
	cmdUtil.EnsureConfigFileExists(newConfFile)

	err = config.WriteViperConfig(newConfFile, sourceConfig)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}
	srcProfileDirs := cmdState.GetMinishiftDirsStructure(constants.GetProfileHomeDir(srcProfile))

	// Removing the newly created addon directory as filehelper.CopyDir does not expect the directory to be present.
	err = os.Remove(newInstanceDirs.Addons)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	// Copying add-ons from source profile to the new profile
	err = filehelper.CopyDir(srcProfileDirs.Addons, newInstanceDirs.Addons)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Copying add-ons to profile '%s' failed : %s", newProfile, err.Error()))
	}

	newProfileInstanceConfig, err := config.NewInstanceConfig(minishiftConstants.GetProfileInstanceConfigPath(newProfile))
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating config for profile '%s' : %s", newProfile, err.Error()))
	}

	srcProfileInstanceConfig, err := config.NewInstanceConfig(minishiftConstants.GetProfileInstanceConfigPath(srcProfile))
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating config for source profile '%s' : %s", srcProfile, err.Error()))
	}

	// Copy addon state information from config/minishift.json
	newProfileInstanceConfig.AddonConfig = srcProfileInstanceConfig.AddonConfig
	if err != newProfileInstanceConfig.Write() {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error copying add-on config to profile '%s': %s", newProfile, err.Error()))
	}

	fmt.Println(fmt.Sprintf("Profile '%s' is created successfully using configs from profile '%s'", newProfile, srcProfile))
}

func validateCopyProfileCmd(args []string) {
	missingArgumentMessage := "Source and a target profile (new profile) name must be provided. Run 'minishift profile list' for a list of existing profiles."
	invalidNameMessage := "Profile names must consist of alphanumeric characters only."
	if len(args) < 2 {
		atexit.ExitWithMessage(1, missingArgumentMessage)
	}

	if !cmdUtil.IsValidProfileName(args[1]) {
		atexit.ExitWithMessage(1, invalidNameMessage)
	}
}

func init() {
	ProfileCmd.AddCommand(profileCopyCmd)
}
