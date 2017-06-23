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

var profileResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "resets the active profile",
	Long:  "resets the active profile. After running this command there will not be any active profile",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			atexit.ExitWithMessage(1, extraArgumentError)
		}

		err := profileActions.ResetActiveProfile()
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		} else {
			fmt.Println("Profile is reset successfully")
		}
	},
}

func init() {
	ProfileCmd.AddCommand(profileResetCmd)
}
