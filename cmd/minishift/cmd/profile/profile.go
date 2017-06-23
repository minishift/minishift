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
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ProfileCmd = &cobra.Command{
	Use:     "profile SUBCOMMAND [flags]",
	Aliases: []string{"instance"},
	Short:   "Manages Minishift profiles.",
	Long:    "Manages Minishift profile. Use the sub-commands to list, set profile",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

//Checks viper or allinstanceconfg to say if profile is defined.
func ProfileExists() bool {
	if viper.GetString("profile") != "" {
		return true
	} else if minishiftConfig.AllInstancesConfig != nil {
		if profileActions.GetActiveProfile() != "" {
			return true
		}
	}
	return false
}

//Returns the machine name after taking in to account if --profile or
//profile information available in allinstance config
func GetProfileName() string {
	activeProfile := profileActions.GetActiveProfile()
	if viper.GetString("profile") != "" {
		return viper.GetString("profile")
	} else if activeProfile != "" {
		return activeProfile
	} else {
		return ""
	}
}
