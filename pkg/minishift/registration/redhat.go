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
	"strings"

	"github.com/docker/machine/libmachine/provision"
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

func (registrator *RedHatRegistrator) CompatibleWithDistribution(osReleaseInfo *provision.OsRelease) bool {
	if osReleaseInfo.ID != "rhel" {
		return false
	}
	if _, err := registrator.SSHCommand("sudo -E subscription-manager version"); err != nil {
		return false
	} else {
		return true
	}
}

func (registrator *RedHatRegistrator) Register(param *RegistrationParameters) error {
	if output, err := registrator.SSHCommand("sudo -E subscription-manager version"); err != nil {
		return err
	} else {
		if strings.Contains(output, "not registered") {
			subscriptionCommand := fmt.Sprintf("sudo -E subscription-manager register --auto-attach "+
				"--username %s "+
				"--password '%s' ", param.Username, param.Password)
			if _, err := registrator.SSHCommand(subscriptionCommand); err != nil {
				return err
			}
		}
	}
	return nil
}

func (registrator *RedHatRegistrator) Unregister(param *RegistrationParameters) error {
	if output, err := registrator.SSHCommand("sudo -E subscription-manager version"); err != nil {
		return err
	} else {
		if !strings.Contains(output, "not registered") {
			if _, err := registrator.SSHCommand(
				"sudo -E subscription-manager unregister"); err != nil {
				return err
			}
		}
	}
	return nil
}
