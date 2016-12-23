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
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/drivers"
)


type RegistrationParametersStruct struct {
	Username      string
	Password      string
	Proxy         string
	Proxyuser     string
	Proxypassword string
}

// Register host VM
func RegisterHostVM(host *host.Host, param *RegistrationParametersStruct) error {
	commander := provision.GenericSSHCommander{Driver: host.Driver}
	registrator, err := DetectRegistrator(commander)
	if err == ErrDetectionFailed {
		fmt.Println("Distribution doesn't support registration")
	} else if err != nil {
		fmt.Errorf("Error running registration: %s", err)
		return fmt.Errorf("Error running registration: %s", err)
	}

	if registrator != nil {
		if param.Username == "" || param.Password == "" {
			return fmt.Errorf("Either MINISHIFT_USERNAME/MINISHIFT_PASSWORD not set as Environment" +
				" variable or username/password not provided using --username/--password\n")
		}
		if err := registrator.Register(param); err != nil {
			return fmt.Errorf("Error running registration: %s", err)
		}
	}
	return nil
}

// Unregister host VM
func UnregisterHostVM(host *host.Host, param *RegistrationParametersStruct) error {
	commander := provision.GenericSSHCommander{Driver: host.Driver}
	registrator, err := DetectRegistrator(commander)

	if err == ErrDetectionFailed {
		fmt.Println("Distribution doesn't support unregistration")
	} else if err != drivers.ErrHostIsNotRunning { } else if err != nil {
		return fmt.Errorf("Error during unregistration: %s", err)

	}

	if registrator != nil {
		fmt.Println("Unregistering Distribution")
		if err := registrator.Unregister(param); err != nil {
			return fmt.Errorf("Error during unregistration: %s", err)
		}
	}
	return nil
}


