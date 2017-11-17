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

package cluster

import (
	"path/filepath"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minikube/tests"
	"io/ioutil"
	"os"
)

func TestRemoteBoot2DockerURL(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-tmp-test-dir-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	var machineConfig = MachineConfig{
		MinikubeISO: "http://github.com/fake/boot2docker.iso",
		ISOCacheDir: filepath.Join(testDir, "iso"),
	}

	isoPath := filepath.Join(testDir, "iso", "unnamed", filepath.Base(machineConfig.MinikubeISO))
	expectedURL := "file://" + filepath.ToSlash(isoPath)
	url := machineConfig.GetISOFileURI()

	if url != expectedURL {
		t.Fatalf("Expected URL : %s, Got : %s", expectedURL, url)
	}
}

func TestLocalBoot2DockerURL(t *testing.T) {
	isoPath := filepath.Join(constants.Minipath, "cache", "iso", "boot2docker.iso")
	localISOUrl := "file://" + filepath.ToSlash(isoPath)

	machineConfig := MachineConfig{
		MinikubeISO: localISOUrl,
	}

	url := machineConfig.GetISOFileURI()

	if url != localISOUrl {
		t.Fatalf("Expected URL : %s", localISOUrl)
	}
}

func TestFollowLogsFlag(t *testing.T) {
	api := tests.NewMockAPI()

	s, _ := tests.NewSSHServer()
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
	api.Hosts[constants.MachineName] = &host.Host{Driver: d}

	testCases := []struct {
		expectedCommand string
		follow          bool
	}{
		{
			expectedCommand: logsCmd,
			follow:          false,
		},
		{
			expectedCommand: logsCmdFollow,
			follow:          true,
		},
	}

	for _, test := range testCases {
		t.Run(test.expectedCommand, func(t *testing.T) {
			if _, err = GetHostLogs(api, test.follow); err != nil {
				t.Errorf("Error getting logs of the running OpenShift cluster: %s", err)
			}
			if _, ok := s.Commands[test.expectedCommand]; !ok {
				t.Errorf("Expected command %s to run but did not.", test.expectedCommand)
			}
		})
	}
}
