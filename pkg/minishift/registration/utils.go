/*
Copyright (C) 2016 Red Hat, Inc.

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

	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/provision"
)

type RegistrationParameters struct {
	Username               string
	Password               string
	IsTtySupported         bool
	GetUsernameInteractive func(message string) string
	GetPasswordInteractive func(message string) string
	GetPasswordKeyring     func(username string) (string, error)
	SetPasswordKeyring     func(username, password string) error
}

// Detect supported Registrator
func NeedRegistration(host *host.Host) (bool, error) {
	commander := provision.GenericSSHCommander{Driver: host.Driver}
	_, supportRegistration, err := DetectRegistrator(commander)
	return supportRegistration, err
}

type RegistrationHostActionFunc func(param *RegistrationParameters) error

func doRegistrationHostAction(actionMessage string, registrationAction RegistrationHostActionFunc, param *RegistrationParameters) (bool, error) {
	fmt.Println(actionMessage)
	if err := registrationAction(param); err != nil {
		// error occured during action
		return false, err
	}

	// was successful
	return true, nil
}

// Register host VM
func RegisterHostVM(host *host.Host, param *RegistrationParameters) (bool, error) {
	commander := provision.GenericSSHCommander{Driver: host.Driver}
	registrator, supportRegistration, err := DetectRegistrator(commander)
	if !supportRegistration {
		log.Debug("Distribution doesn't support registration")
	}
	if err != nil && err != ErrDetectionFailed {
		return supportRegistration, err
	}
	if err != nil && err != ErrDetectionFailed {
		// failed
		return supportRegistration, err
	}
	if err == ErrDetectionFailed || registrator == nil {
		// does not support registration
		return supportRegistration, nil
	}

	return doRegistrationHostAction(
		"-- Registering machine using subscription-manager",
		registrator.Register,
		param)
}

// Unregister host VM
func UnregisterHostVM(host *host.Host, param *RegistrationParameters) (bool, error) {
	commander := provision.GenericSSHCommander{Driver: host.Driver}
	registrator, supportRegistration, err := DetectRegistrator(commander)
	if !supportRegistration {
		log.Debug("Distribution doesn't support unregistration")
	}
	if err != nil && err != ErrDetectionFailed {
		// failed
		return supportRegistration, err
	}
	if err == ErrDetectionFailed || registrator == nil {
		// does not support unregistration
		return supportRegistration, nil
	}

	return doRegistrationHostAction(
		"Unregistering machine",
		registrator.Unregister,
		param)
}
