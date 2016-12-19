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

func TestRedhatRegistratorCompatibleWithDistribution(t *testing.T) {
	s, _ := tests.NewSSHServer()
	s.CommandToOutput = make(map[string]string)
	s.CommandToOutput["sudo subscription-manager version"] = `server type: This system is currently not registered.
subscription management server: 0.9.51.11-1
subscription management rules: 5.15
subscription-manager: 1.17.15-1.el7
python-rhsm: 1.17.9-1.el7
`
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

	registrator := NewRedhatRegistrator(d)
	info := &provision.OsRelease{
		Name:      "Red Hat Enterprise Linux Server",
		ID: "rhel",
		VersionID: "7.3",
	}
	if !registrator.CompatibleWithDistribution(info) {
		t.Fatal("Registration capability should be in the Distribution")
	}
}

func TestRedhatRegistratorRegister(t *testing.T) {
	s, _ := tests.NewSSHServer()
	s.CommandToOutput = make(map[string]string)
	s.CommandToOutput["sudo subscription-manager version"] = `server type: This system is currently not registered.
subscription management server: 0.9.51.11-1
subscription management rules: 5.15
subscription-manager: 1.17.15-1.el7
python-rhsm: 1.17.9-1.el7
`
	s.CommandToOutput["sudo subscription-manager register --auto-attach --username foo --password foo"] = `
	The system has been registered with ID: 2af78473-n5th-ksa1-3748-983748c77da`
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

	registrator := NewRedhatRegistrator(d)
	m := make(map[string]string)
	m["username"] = "foo"
	m["password"] = "foo"
	if err := registrator.Register(m); err != nil {
		t.Fatal("Distribution should able to register")
	}
}

func TestRedhatRegistratorUnregister(t *testing.T) {
	s, _ := tests.NewSSHServer()
	s.CommandToOutput = make(map[string]string)
	s.CommandToOutput["sudo subscription-manager version"] = `server type: This system is currently not registered.
subscription management server: 0.9.51.11-1
subscription management rules: 5.15
subscription-manager: 1.17.15-1.el7
python-rhsm: 1.17.9-1.el7
`
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

	registrator := NewRedhatRegistrator(d)
	if err := registrator.Unregister(); err != nil {
		t.Fatal("Distribution shouldn't able to unregister")
	}
}
