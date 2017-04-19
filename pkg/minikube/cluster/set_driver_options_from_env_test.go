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

package cluster

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/minishift/minishift/pkg/minikube/tests"
	"github.com/minishift/minishift/pkg/testing/cli"
	"os"
	"testing"
)

func TestSetDriverOptionsFromEnvironment(t *testing.T) {
	testDir := cli.SetupTmpMinishiftHome(t)
	tee := cli.CreateTee(t, true)
	defer cli.TearDown(testDir, tee)

	d := &tests.MockDriver{
		BaseDriver: drivers.BaseDriver{},
	}
	expectedURL := "/foo/bar/boot2docker.iso"
	os.Setenv("TEST_BOOT2DOCKER_URL", expectedURL)
	defer os.Unsetenv("TEST_BOOT2DOCKER_URL")

	setDriverOptionsFromEnvironment(d)
	if d.Boot2DockerURL != expectedURL {
		t.Errorf("Expected %s but got %s", expectedURL, d.Boot2DockerURL)
	}
}

func TestSetWrongDriverOptionFromEnvironment(t *testing.T) {
	testDir := cli.SetupTmpMinishiftHome(t)
	tee := cli.CreateTee(t, true)
	defer cli.TearDown(testDir, tee)

	d := &tests.MockDriver{
		BaseDriver: drivers.BaseDriver{},
	}
	expectedURL := "/foo/bar/boot2docker.iso"
	os.Setenv("WRONG_BOOT2DOCKER_URL", expectedURL)
	defer os.Unsetenv("WRONG_BOOT2DOCKER_URL")

	setDriverOptionsFromEnvironment(d)
	if d.Boot2DockerURL == expectedURL {
		t.Errorf("Expected %s to not equal to %s", expectedURL, d.Boot2DockerURL)
	}
}
