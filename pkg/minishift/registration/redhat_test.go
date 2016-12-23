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
	"github.com/docker/machine/libmachine/drivers"
	"github.com/minishift/minishift/pkg/minikube/tests"
	"testing"
	"github.com/docker/machine/libmachine/provision"
)

var (
	param = &RegistrationParametersStruct{
		Username: "foo",
		Password: "foo",
	}
)

func TestRedhatRegistratorCompatibleWithDistribution(t *testing.T) {
	s, _ := tests.NewSSHServer()
	s.CommandToOutput = make(map[string]string)
	s.CommandToOutput["sudo subscription-manager version"] = `server type: This system is currently not registered.`
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

	registrator := NewRedhatRegistrator(commander)
	info := &provision.OsRelease{
		Name:      "Red Hat Enterprise Linux Server",
		ID: 	   "rhel",
		VersionID: "7.3",
	}
	if !registrator.CompatibleWithDistribution(info) {
		t.Fatal("Registration capability should be in the Distribution")
	}
}

func TestRedhatRegistratorRegister(t *testing.T) {
	s, _ := tests.NewSSHServer()
	s.CommandToOutput = make(map[string]string)
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
	registrator := NewRedhatRegistrator(commander)
	s.CommandToOutput["sudo subscription-manager version"] = `server type: RedHat Subscription Management`
	if err := registrator.Register(param); err != nil {
		t.Fatal("Distribution shouldn't have to register")
	} else {
		cmd := "sudo subscription-manager register --auto-attach --username foo --password foo"
		if _, ok := s.Commands[cmd]; ok{
			t.Fatalf("Expected command: %s", cmd)
		}
	}
	s.CommandToOutput["sudo subscription-manager version"] = `server type: This system is currently not registered.`
	if err := registrator.Register(param); err != nil {
		t.Fatal("Distribution should able to register")
	} else {
		cmd := "sudo subscription-manager register --auto-attach --username foo --password foo "
		if _, ok := s.Commands[cmd]; !ok{
			t.Fatalf("Expected command: %s", cmd)
		}
	}

}

func TestRedhatRegistratorUnregister(t *testing.T) {
	s, _ := tests.NewSSHServer()
	s.CommandToOutput = make(map[string]string)
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
	registrator := NewRedhatRegistrator(commander)
	s.CommandToOutput["sudo subscription-manager version"] = `server type: This system is currently not registered.`
	if err := registrator.Unregister(param); err != nil {
		t.Fatal("Distribution shouldn't able to unregister")
	} else {
		cmd := "sudo subscription-manager unregister"
		if _, ok := s.Commands[cmd]; ok{
			t.Fatalf("Expected command: %s", cmd)
		}
	}
	s.CommandToOutput["sudo subscription-manager version"] = `server type: RedHat Subscription Management`
	if err := registrator.Unregister(param); err != nil {
		t.Fatal("Distribution should able to unregister")
	} else {
		cmd := "sudo subscription-manager unregister"
		if _, ok := s.Commands[cmd]; !ok{
			t.Fatalf("Expected command: %s", cmd)
		}
	}
}
