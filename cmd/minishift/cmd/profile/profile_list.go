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
	"text/tabwriter"

	"github.com/minishift/minishift/pkg/minikube/cluster"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/pkg/minikube/constants"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists profiles.",
	Long:  "Lists the existing profiles.",
	Run: func(cmd *cobra.Command, args []string) {
		profiles := profileActions.GetProfileMap()
		displayProfiles(profiles)
	},
}

func displayProfiles(profiles map[string]bool) {
	display := new(tabwriter.Writer)
	display.Init(os.Stdout, 0, 8, 0, '\t', 0)

	for name, isActive := range profiles {
		vmStatus := getVmStatus(name)
		if isActive {
			fmt.Fprintln(display, fmt.Sprintf("- %s\t%s\t(Active profile)", name, vmStatus))
		} else {
			fmt.Fprintln(display, fmt.Sprintf("- %s\t%s", name, vmStatus))
		}
	}
	display.Flush()
}

func getVmStatus(name string) string {
	constants.ProfileName = name
	constants.MachineName = constants.ProfileName
	constants.Minipath = constants.GetProfileHomeDir()
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()
	status, err := cluster.GetHostStatus(api)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting cluster status: %s", err.Error()))
	}
	return status
}
func init() {
	ProfileCmd.AddCommand(profileListCmd)
}
