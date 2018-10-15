/*
Copyright (C) 2018 Red Hat, Inc.

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

package cmd

import (
	"fmt"
	"github.com/minishift/minishift/pkg/minishift/setup/hypervisor"
	"github.com/minishift/minishift/pkg/minishift/shell/powershell"
	"github.com/minishift/minishift/pkg/util/os"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

const (
	assumeYesFlag = "yes"
)

var (
	assumeYes bool
)

// setupCmd represents the command to setup Minishift pre-requisites
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configures pre-requisites for Minishift on the host machine",
	Long:  `Configures pre-requisites for Minishift on the host machine`,
	Run:   runSetup,
}

func init() {
	setupCmd.Flags().BoolVarP(&assumeYes, assumeYesFlag, "y", false, "Automatically consider 'yes' for all questions")
	RootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) {
	if os.CurrentOS() == "windows" {
		if !powershell.IsAdmin() {
			atexit.ExitWithMessage(1, "Run 'minishift setup' as an administrator.")
		}
		if err := hypervisor.CheckHypervisorAvailable(); err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		if err := hypervisor.CheckAndConfigureHypervisor(); err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		fmt.Println("Pre-requisites are ready.\nRun following commands to start your minishift instance:\n" +
			" minishift config set hyperv-virtual-switch minishift-external\n minishift start\n" +
			"Note: Above two commands doesn't need to be run in Administrator mode.")
	} else {
		fmt.Println("The implementation of this command is not available for this operating system.\n" +
			"Please continue with setting pre-requisites through manual process.")
	}

	fmt.Println("\n\nConsider checking getting started guide at https://docs.okd.io/latest/minishift/getting-started/index.html")
}
