/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"time"

	"github.com/minishift/minishift/pkg/minishift/timezone"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var (
	timezoneCmd = &cobra.Command{
		Use:    "timezone",
		Short:  "Set timezone for running Minishift Instance",
		Long:   `Set timezone for running Minishift Instance, provided timezone should be valid.`,
		Run:    runTimezone,
		Hidden: true,
	}
	set  string
	list bool
)

func runTimezone(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer api.Close()

	// if VM does not exist, exit with error
	util.ExitIfUndefined(api, constants.MachineName)

	hostVm, err := api.Load(constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	if list {
		out, err := timezone.GetTimeZoneList(hostVm)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
		fmt.Println(out)
		return
	}

	if set == "" {
		out, err := timezone.GetTimeZone(hostVm)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
		fmt.Println(out)
		return
	}
	_, err = time.LoadLocation(set)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("%s is not a vaild timezone: %s", set, err.Error()))
	}

	fmt.Printf("Setting timezone for Instance ... ")
	if err := timezone.UpdateTimeZone(hostVm, set); err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}
	fmt.Println("OK")
}

func init() {
	timezoneCmd.Flags().StringVarP(&set, "set", "s", "", "Set the provided timezone.")
	timezoneCmd.Flags().BoolVarP(&list, "list", "l", false, "List of the available timezone to Minishift Instance")
	RootCmd.AddCommand(timezoneCmd)
}
