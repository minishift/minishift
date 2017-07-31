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

package cluster

import (
	"fmt"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/registration"
)

var (
	RegistrationParameters = new(registration.RegistrationParameters)
)

type RegistrationActionFunc func(host *host.Host, param *registration.RegistrationParameters) (bool, error)

func doRegistrationAction(api libmachine.API, registrationAction RegistrationActionFunc) (bool, error) {
	host, err := api.Load(constants.MachineName)
	if err != nil {
		return false, err
	}

	if !drivers.MachineInState(host.Driver, state.Running)() {
		return false, fmt.Errorf("Unable to perform registration action. VM not in running state")
	}

	return registrationAction(host, RegistrationParameters)
}

// Register returns true if successfully registered
func Register(api libmachine.API) (bool, error) {
	return doRegistrationAction(api, registration.RegisterHostVM)
}

// UnRegister returns true if successfully unregistered
func UnRegister(api libmachine.API) (bool, error) {
	return doRegistrationAction(api, registration.UnregisterHostVM)
}
