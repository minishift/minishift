/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package sshutil

import (
	"bytes"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"

	"github.com/minishift/minishift/pkg/minikube/tests"
)

func TestNewSSHClient(t *testing.T) {
	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	assert.NoError(t, err, "Error starting the ssh server")
	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	c, err := NewSSHClient(d)
	assert.NoError(t, err, "Error starting ssh client")

	cmd := "foo"
	RunCommand(c, cmd)
	assert.True(t, s.Connected, "Error connecting to ssh")

	_, ok := s.Commands[cmd]
	assert.True(t, ok, "Expected command: %s", cmd)

}

func TestNewSSHHost(t *testing.T) {
	sshKeyPath := "mypath"
	ip := "localhost"
	user := "myuser"
	d := tests.MockDriver{
		BaseDriver: drivers.BaseDriver{
			IPAddress:  ip,
			SSHUser:    user,
			SSHKeyPath: sshKeyPath,
		},
	}

	h, err := newSSHHost(&d)
	assert.NoError(t, err, "Error creating host")

	assert.Equal(t, sshKeyPath, h.SSHKeyPath)
	assert.Equal(t, user, h.Username)
	assert.Equal(t, ip, h.IP)
}

func TestNewSSHHostError(t *testing.T) {
	d := tests.MockDriver{HostError: true}

	_, err := newSSHHost(&d)
	assert.Error(t, err, "Expected error creating host")
}

func TestTransfer(t *testing.T) {
	s, _ := tests.NewSSHServer()
	port, err := s.Start()
	assert.NoError(t, err, "Error starting ssh server")
	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	c, err := NewSSHClient(d)
	assert.NoError(t, err, "Error starting ssh client")

	dest := "bar"
	contents := []byte("testcontents")
	err = Transfer(bytes.NewReader(contents), int64(len(contents)), "/tmp", dest, "0777", c)
	assert.NoError(t, err, "Error transferring bytes")
}
