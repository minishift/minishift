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
	"strings"
)

func init() {
	Register("Redhat", &RegisteredRegistrator{
		New: NewRedhatRegistrator,
	})
}

func NewRedhatRegistrator(c provision.SSHCommander) Registrator {
	return &RedhatRegistrator{
		SSHCommander: c,
	}
}

type RedhatRegistrator struct {
	provision.SSHCommander
}

func (registrator *RedhatRegistrator) CompatibleWithDistribution(osReleaseInfo *provision.OsRelease) bool {
	if osReleaseInfo.ID != "rhel" {
		return false
	}
	if _, err := registrator.SSHCommand("sudo subscription-manager version"); err != nil {
		return false
	} else {
		return true
	}
}

func (registrator *RedhatRegistrator) Register(param *RegistrationParametersStruct) error {
	if output, err := registrator.SSHCommand("sudo subscription-manager version"); err != nil {
		return err
	} else {
		if strings.Contains(output, "not registered") {
			subscriptionCommand := fmt.Sprintf("sudo subscription-manager register --auto-attach " +
							"--username %s " +
							"--password %s ", param.Username, param.Password)
			if _, err := registrator.SSHCommand(subscriptionCommand); err != nil {
				return err
			}
		}
	}
	return nil
}

func (registrator *RedhatRegistrator) Unregister(param *RegistrationParametersStruct) error {
	if output, err := registrator.SSHCommand("sudo subscription-manager version"); err != nil {
		return err
	} else {
		if !strings.Contains(output, "not registered") {
			if _, err := registrator.SSHCommand(
				"sudo subscription-manager unregister"); err != nil {
				return err
			}
		}
	}
	return nil
}
