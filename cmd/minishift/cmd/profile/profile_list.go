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
	"sort"
	"text/tabwriter"

	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists profiles.",
	Long:  "Lists the existing profiles.",
	Run: func(cmd *cobra.Command, args []string) {
		profiles := profileActions.GetProfileList()
		displayProfiles(profiles)
	},
}

func displayProfiles(profiles []string) {
	display := new(tabwriter.Writer)
	display.Init(os.Stdout, 0, 8, 2, '\t', 0)

	activeProfile := profileActions.GetActiveProfile()
	sort.Strings(profiles)
	for _, profile := range profiles {
		vmStatus := cmdUtil.GetVMStatus(profile)
		if profile == activeProfile {
			fmt.Fprintln(display, fmt.Sprintf("- %s\t%s\t%s", activeProfile, vmStatus, "(Active)"))
		} else {
			fmt.Fprintln(display, fmt.Sprintf("- %s\t%s", profile, vmStatus))
		}
	}
	display.Flush()
}

func init() {
	ProfileCmd.AddCommand(profileListCmd)
}
