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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/util"
	minishiftStrings "github.com/minishift/minishift/pkg/util/strings"
)

const (
	RHSMName = "   Red Hat Developers or Red Hat Subscription Management (RHSM)"
)

func init() {
	Register("Redhat", &RegisteredRegistrator{
		New: NewRedHatRegistrator,
	})
}

func NewRedHatRegistrator(c provision.SSHCommander) Registrator {
	return &RedHatRegistrator{
		SSHCommander: c,
	}
}

type RedHatRegistrator struct {
	provision.SSHCommander
}

// CompatibleWithDistribution returns true if system supports registration with RHSM
func (registrator *RedHatRegistrator) CompatibleWithDistribution(osReleaseInfo *provision.OsRelease) bool {
	if osReleaseInfo.ID != "rhel" {
		return false
	}
	if _, err := registrator.SSHCommand("sudo -E subscription-manager"); err != nil {
		return false
	} else {
		return true
	}
}

// Register attempts to register the system with RHSM
func (registrator *RedHatRegistrator) Register(param *RegistrationParameters) error {
	progressDots := make(chan bool)

	if isRegistered, err := registrator.isRegistered(); !isRegistered && err == nil {
		for i := 1; i < 4; i++ {
			// request username (disallow empty value)
			if param.Username == "" {
				for param.Username == "" {
					param.Username = param.GetUsernameInteractive(RHSMName + " username")
				}
			}
			// request password (disallow empty value)
			if param.Password == "" {
				for param.Password == "" {
					param.Password = param.GetPasswordInteractive(RHSMName + " password")
				}
			}

			// prepare subscription command
			subscriptionCommand := fmt.Sprintf("sudo -E subscription-manager register --auto-attach "+
				"--username %s "+
				"--password '%s' ",
				param.Username,
				minishiftStrings.EscapeSingleQuote(param.Password))

			fmt.Print("   Registration in progress ")
			util.StartProgressDots(progressDots)
			startTime := time.Now()
			// start timed SSH command to register
			_, err = registrator.SSHCommand(subscriptionCommand)
			util.StopProgressDots(progressDots)
			if err == nil {
				fmt.Print(" OK ")
			} else {
				fmt.Print(" FAIL ")
			}
			fmt.Printf("[%s]\n", util.TimeElapsed(startTime, true))

			if err == nil {
				return nil
			}

			// general error when registration fails
			if strings.Contains(err.Error(), "Invalid username or password") {
				fmt.Println("   Invalid username or password. Retry:", i)
			} else {
				return err
			}

			// always reset credentials
			param.Username = ""
			param.Password = ""
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Unregister attempts to unregister the system from RHSM
func (registrator *RedHatRegistrator) Unregister(param *RegistrationParameters) error {
	if isRegistered, err := registrator.isRegistered(); isRegistered {
		if _, err := registrator.SSHCommand(
			"sudo -E subscription-manager unregister"); err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

// isRegistered returns registration state of RHSM or errors when undetermined
func (registrator *RedHatRegistrator) isRegistered() (bool, error) {
	if output, err := registrator.SSHCommand("sudo -E subscription-manager version"); err != nil {
		return false, err
	} else {
		if !strings.Contains(output, "not registered") {
			return true, nil
		}
		return false, nil
	}

	return false, errors.New("Unable to determine registration state")
}
