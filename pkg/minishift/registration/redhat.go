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
	"github.com/docker/machine/libmachine/provision"
	"strings"
)

const (
	RHSMName       = "   Red Hat Developers or Red Hat Subscription Management (RHSM)"
	RedHatRegistry = "registry.redhat.io"
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
	}
	return true
}

// Register attempts to register the system with RHSM and registry.redhat.io
func (registrator *RedHatRegistrator) Register(param *RegistrationParameters) error {
	if isRegistered, err := registrator.IsRegistered(); !isRegistered && err == nil {
		err := RSHMRegisterAndLoginToRedhatRegistry(registrator.SSHCommander, param, true)
		if err != nil {
			return errors.New(redactPassword(err.Error()))
		}
	}
	return nil
}

// Unregister attempts to unregister the system from RHSM
func (registrator *RedHatRegistrator) Unregister(param *RegistrationParameters) error {
	if isRegistered, err := registrator.IsRegistered(); isRegistered {
		if _, err := registrator.SSHCommand(
			"sudo -E subscription-manager unregister"); err != nil {
			return err
		}
		if _, err := registrator.SSHCommand(
			"sudo -E subscription-manager clean"); err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

// isRegistered returns registration state of RHSM or errors when undetermined
func (registrator *RedHatRegistrator) IsRegistered() (bool, error) {
	if output, err := registrator.SSHCommand("sudo -E subscription-manager list"); err != nil {
		return false, err
	} else {
		if !strings.Contains(output, "Unknown") {
			return true, nil
		}
		return false, nil
	}
}
