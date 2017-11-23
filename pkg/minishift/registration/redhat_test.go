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
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/minikube/tests"
	"github.com/stretchr/testify/assert"
)

var (
	param = &RegistrationParameters{
		Username: "foo",
		Password: "foo",
	}
	expectedCMDRegistration = fmt.Sprintf("sudo -E subscription-manager register --auto-attach --username %s --password '%s' ",
		param.Username, param.Password)
	expectedCMDUnregistration = "sudo -E subscription-manager unregister"
)

func setup(t *testing.T) (registrator Registrator) {
	s, _ := tests.NewSSHServer()
	s.CommandToOutput = make(map[string]string)
	s.CommandToOutput["sudo -E subscription-manager version"] = `server type: This system is currently not registered.`
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}
	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}

	commander := provision.GenericSSHCommander{Driver: d}

	registrator = NewRedHatRegistrator(commander)
	return registrator
}

func TestRedHatRegistratorCompatibleWithDistribution(t *testing.T) {
	registrator := setup(t)
	info := &provision.OsRelease{
		Name:      "Red Hat Enterprise Linux Server",
		ID:        "rhel",
		VersionID: "7.3",
	}
	assert.True(t, registrator.CompatibleWithDistribution(info), "Registration capability should be in the Distribution")
}

func TestRedHatRegistratorNotCompatibleWithDistribution(t *testing.T) {
	registrator := setup(t)
	info := &provision.OsRelease{
		Name:      "CentOS",
		ID:        "centos",
		VersionID: "7.3",
	}
	assert.False(t, registrator.CompatibleWithDistribution(info), "Registration capability shouldn't be in the Distribution")
}

func TestRedHatRegistratorRegister(t *testing.T) {
	s, _ := tests.NewSSHServer()
	s.CommandToOutput = make(map[string]string)
	port, err := s.Start()
	assert.NoError(t, err, "Error starting ssh server")
	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	commander := provision.GenericSSHCommander{Driver: d}
	registrator := NewRedHatRegistrator(commander)

	s.CommandToOutput["sudo -E subscription-manager version"] = `server type: This system is currently not registered.`
	err = registrator.Register(param)
	assert.NoError(t, err, "Distribution should be able to register")
	_, ok := s.Commands[expectedCMDRegistration]
	assert.True(t, ok, "Expected command :%s", expectedCMDRegistration)

}

func TestRedHatRegistratorUnregister(t *testing.T) {
	s, _ := tests.NewSSHServer()
	s.CommandToOutput = make(map[string]string)
	port, err := s.Start()
	assert.NoError(t, err, "Error starting ssh server")
	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	commander := provision.GenericSSHCommander{Driver: d}
	registrator := NewRedHatRegistrator(commander)

	s.CommandToOutput["sudo -E subscription-manager version"] = `server type: RedHat Subscription Management`
	err = registrator.Unregister(param)
	assert.NoError(t, err, "Distribution should be able to unregister")
	_, ok := s.Commands[expectedCMDUnregistration]
	assert.True(t, ok, "Expected command: %s", expectedCMDUnregistration)
}
