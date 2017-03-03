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
	"github.com/pkg/errors"
)

type RegistrationParameters struct {
	Username string
	Password string
}

// Register host VM
func RegisterHostVM(host *host.Host, param *RegistrationParameters) error {
	commander := provision.GenericSSHCommander{Driver: host.Driver}
	registrator, supportRegistration, err := DetectRegistrator(commander)
	if !supportRegistration {
		log.Debug("Distribution doesn't support registration")
	}

	if err != nil && err != ErrDetectionFailed {
		return err
	}

	if registrator != nil {
		fmt.Println("Registering machine using subscription-manager")
		if param.Username == "" || param.Password == "" {
			return errors.New("This virtual machine requires registration. " +
				"Credentials must either be passed via the environment variables " +
				"MINISHIFT_USERNAME and MINISHIFT_PASSWORD " +
				" or the --username and --password flags\n")
		}

		if err := registrator.Register(param); err != nil {
			return err
		}
	}
	return nil
}

// Unregister host VM
func UnregisterHostVM(host *host.Host, param *RegistrationParameters) error {
	commander := provision.GenericSSHCommander{Driver: host.Driver}
	registrator, supportUnregistration, err := DetectRegistrator(commander)

	if !supportUnregistration {
		log.Debug("Distribution doesn't support unregistration")
	}

	if err != nil && err != ErrDetectionFailed {
		return err
	}

	if registrator != nil {
		fmt.Println("Unregistering machine")
		if err := registrator.Unregister(param); err != nil {
			return err
		}
	}
	return nil
}
