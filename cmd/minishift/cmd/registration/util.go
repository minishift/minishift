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

package registration

import (
	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
	"github.com/minishift/minishift/pkg/minishift/cluster"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	pkgUtil "github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

var (
	SkipRegistration   bool
	SkipUnRegistration bool
)

func isRegistered() bool {
	return minishiftConfig.InstanceConfig.IsRegistered
}

func setRegistered(state bool) {
	minishiftConfig.InstanceConfig.IsRegistered = state
	minishiftConfig.InstanceConfig.Write()
}

func RegisterHost(libMachineClient *libmachine.Client) {
	if SkipRegistration {
		log.Debug("Skipping registration due to enabled '--skip-registration' flag")
		return
	}

	if wasSuccessful, err := cluster.Register(libMachineClient); err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error to register VM: %v", err))
	} else {
		// else, we set the IsRegistered state according to result
		setRegistered(wasSuccessful)
	}
}

func UnregisterHost(api libmachine.API, allowToSkipUnregistration bool) {
	if allowToSkipUnregistration && SkipUnRegistration {
		log.Debug("Skipping unregistration due to enabled '--skip-unregistration' flag")
		return
	}

	if !isRegistered() {
		return
	}

	if wasSuccessful, err := cluster.UnRegister(api); err != nil || !wasSuccessful {
		handleFailedRegistrationAction()
	}

	// if continued, we remove the IsRegistered state
	setRegistered(false)
}

func handleFailedRegistrationAction() {
	hasConfirmed := pkgUtil.AskForConfirmation("Current Minishift VM is registered, but unregistration failed.")
	if !hasConfirmed {
		atexit.ExitWithMessage(0, fmt.Sprintln("Aborted."))
	}
}
