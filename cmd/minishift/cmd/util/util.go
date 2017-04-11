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

package util

import (
	"fmt"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

func VMExists(client *libmachine.Client, machineName string) bool {
	exists, err := client.Exists(machineName)
	if err != nil {
		atexit.ExitWithMessage(1, "Unable to determine state of Minishift VM.")
	}
	return exists
}

func ExitIfUndefined(client *libmachine.Client, machineName string) {
	exists := VMExists(client, machineName)
	if !exists {
		atexit.ExitWithMessage(0, fmt.Sprintf("There is currently no '%s' VM defined.", machineName))
	}
}

func IsHostRunning(driver drivers.Driver) bool {
	return drivers.MachineInState(driver, state.Running)()
}

func ExitIfNotRunning(driver drivers.Driver) {
	running := IsHostRunning(driver)
	if !running {
		atexit.ExitWithMessage(0, "Minishift VM is not running.")
	}
}
