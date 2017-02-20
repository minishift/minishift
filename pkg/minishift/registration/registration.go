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

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/provision"
)

var (
	ErrDetectionFailed          = errors.New("Distribution type not recognized to Registration")
	registrators                = make(map[string]*RegisteredRegistrator)
	detector           Detector = &StandardRegistrator{}
)

type Detector interface {
	DetectRegistrator(c provision.SSHCommander) (Registrator, bool, error)
}

type StandardRegistrator struct{}

func SetDetector(newDetector Detector) {
	detector = newDetector
}

// Registration defines distribution specific actions
type Registrator interface {
	provision.SSHCommander

	// Register
	Register(param *RegistrationParameters) error

	// Return the auth options used to configure remote connection for the daemon.
	Unregister(param *RegistrationParameters) error

	// Figure out whether this is a matching registrar
	CompatibleWithDistribution(osReleaseInfo *provision.OsRelease) bool
}

// RegisteredRegistrator creates a new registrator
type RegisteredRegistrator struct {
	New func(c provision.SSHCommander) Registrator
}

func Register(name string, r *RegisteredRegistrator) {
	registrators[name] = r
}

func DetectRegistrator(c provision.SSHCommander) (Registrator, bool, error) {
	return detector.DetectRegistrator(c)
}

func (detector StandardRegistrator) DetectRegistrator(c provision.SSHCommander) (Registrator, bool, error) {
	osReleaseOut, err := c.SSHCommand("sudo cat /etc/os-release")
	if err != nil {
		return nil, false, err
	}
	osReleaseInfo, err := provision.NewOsRelease([]byte(osReleaseOut))
	if err != nil {
		return nil, false, fmt.Errorf("Error parsing /etc/os-release file: %s", err)
	}

	log.Debug("Detecting whether virtual machine is managed by a registration manager...")

	for _, r := range registrators {
		registrator := r.New(c)

		if registrator.CompatibleWithDistribution(osReleaseInfo) {
			log.Debugf("found compatible host")
			return registrator, true, err
		}
	}

	return nil, false, ErrDetectionFailed
}
