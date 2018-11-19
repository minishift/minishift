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

package cluster

import (
	"fmt"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/registration"
)

type RegistryActionFunc func(sshCommander provision.SSHCommander, param *registration.RegistrationParameters, IsRHEL bool) error

func doRegistryAction(api libmachine.API, registryAction RegistryActionFunc) error {
	host, err := api.Load(constants.MachineName)
	if err != nil {
		return err
	}

	if !drivers.MachineInState(host.Driver, state.Running)() {
		return fmt.Errorf("Unable to perform registration action. VM not in running state")
	}

	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}
	return registryAction(sshCommander, RegistrationParameters, false)
}

// LoginToRegistry returns true if successfully logged in to registry.redhat.com
func LoginToRegistry(api libmachine.API) error {
	return doRegistryAction(api, registration.RSHMRegisterAndLoginToRedhatRegistry)
}
