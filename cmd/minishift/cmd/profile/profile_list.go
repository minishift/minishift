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
	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists profiles",
	Long:  "Lists existing profiles for Minishift",
	Run: func(cmd *cobra.Command, args []string) {
		profiles := profileActions.GetProfileNameList()
		if len(profiles) == 0 {
			fmt.Println("There are no profiles defined")
		} else {
			displayProfiles(profiles)
		}
	},
}

//List existing profile information
//TODO: Name, Status (running,stopped etc)
func displayProfiles(profiles []string) {
	for i := range profiles {
		fmt.Println(fmt.Sprintf("- %s", profiles[i]))
	}
}

func init() {
	ProfileCmd.AddCommand(profileListCmd)
}
