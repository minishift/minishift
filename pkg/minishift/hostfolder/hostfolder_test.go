/*
Copyright (C) 2017 Red Hat, Inc.

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

package hostfolder

import (
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/minishift/minishift/pkg/minikube/tests"

	instanceState "github.com/minishift/minishift/pkg/minishift/config"
)

func setupMock() (*tests.SSHServer, *tests.MockDriver) {
	mockSsh, _ := tests.NewSSHServer()
	mockSsh.CommandToOutput = make(map[string]string)
	port, _ := mockSsh.Start()

	mockDriver := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}
	return mockSsh, mockDriver
}

func setupHostFolder() *instanceState.HostFolder {
	return &instanceState.HostFolder{
		Name: "Users",
		Type: "cifs",
		Options: map[string]string{
			"mountpoint": "",
			"uncpath":    "//127.0.0.1/Users",
			"username":   "joe@pillow.us",
			"password":   "am!g@4ever",
			"domain":     "DESKTOP-RHAIMSWIN",
		},
	}
}

func TestHostfolderIsMounted(t *testing.T) {
	mockSsh, mockDriver := setupMock()
	hostfolder := setupHostFolder()

	state := false
	mockSsh.CommandToOutput["if grep -qs /mnt/sda1/Users /proc/mounts; then echo '1'; else echo '0'; fi"] = `0`
	state, _ = isHostfolderMounted(mockDriver, hostfolder)
	if state {
		t.Errorf("Hostfolder error: should have returned 0")
	}

	mockSsh.CommandToOutput["if grep -qs /mnt/sda1/Users /proc/mounts; then echo '1'; else echo '0'; fi"] = `1`
	state, _ = isHostfolderMounted(mockDriver, hostfolder)
	if !state {
		t.Errorf("Hostfolder error: should have returned 1")
	}
}
