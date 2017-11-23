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
	"os"
	"strconv"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/minishift/minishift/pkg/minikube/tests"
	"github.com/stretchr/testify/assert"
)

func Test_driver_options_from_config(t *testing.T) {
	d := &tests.MockDriver{
		BaseDriver: drivers.BaseDriver{},
	}
	supportedFlags := d.GetCreateFlags()

	expectedURL := "/foo/bar/boot2docker.iso"
	explicitConfig := make(map[string]interface{})
	explicitConfig["test-boot2docker-url"] = expectedURL

	driverOptions, err := prepareDriverOptions(supportedFlags, explicitConfig)
	assert.NoError(t, err, "Error preparing driver options ")

	err = d.SetConfigFromFlags(driverOptions)
	assert.NoError(t, err, "Error setting configs from flags")

	assert.Equal(t, expectedURL, d.Boot2DockerURL)
}

func Test_driver_options_from_environment(t *testing.T) {
	d := &tests.MockDriver{
		BaseDriver: drivers.BaseDriver{},
	}
	supportedFlags := d.GetCreateFlags()

	expectedURL := "/foo/bar/boot2docker.iso"
	os.Setenv("TEST_BOOT2DOCKER_URL", expectedURL)
	defer os.Unsetenv("TEST_BOOT2DOCKER_URL")

	driverOptions, err := prepareDriverOptions(supportedFlags, make(map[string]interface{}))
	assert.NoError(t, err, "Error preparing driver options ")

	err = d.SetConfigFromFlags(driverOptions)
	assert.NoError(t, err, "Error setting configs from flags")

	assert.Equal(t, expectedURL, d.Boot2DockerURL)
}

func Test_driver_options_from_environment_int_type(t *testing.T) {
	d := &tests.MockDriver{
		BaseDriver: drivers.BaseDriver{},
	}
	supportedFlags := d.GetCreateFlags()

	expectedCpuCount := 8
	os.Setenv("TEST_CPU_COUNT", strconv.Itoa(expectedCpuCount))
	defer os.Unsetenv("TEST_CPU_COUNT")

	driverOptions, err := prepareDriverOptions(supportedFlags, make(map[string]interface{}))
	assert.NoError(t, err, "Error preparing driver options ")

	actualCount := driverOptions.Int("test-cpu-count")
	assert.Equal(t, expectedCpuCount, actualCount)
}

func Test_driver_options_from_environment_win(t *testing.T) {
	d := &tests.MockDriver{
		BaseDriver: drivers.BaseDriver{},
	}
	supportedFlags := d.GetCreateFlags()

	expectedURL := "/foo/bar/boot2docker.iso"
	os.Setenv("TEST_BOOT2DOCKER_URL", expectedURL)
	defer os.Unsetenv("TEST_BOOT2DOCKER_URL")
	explicitConfig := make(map[string]interface{})
	explicitConfig["test-boot2docker-url"] = "/snafu/boot2docker.iso"

	driverOptions, err := prepareDriverOptions(supportedFlags, explicitConfig)
	assert.NoError(t, err, "Error preparing dirver options")

	err = d.SetConfigFromFlags(driverOptions)
	assert.NoError(t, err, "Error setting configs from flags")

	assert.Equal(t, expectedURL, d.Boot2DockerURL)
}

func Test_wrong_driver_option_does_not_affect_driver_config(t *testing.T) {
	d := &tests.MockDriver{
		BaseDriver: drivers.BaseDriver{},
	}
	supportedFlags := d.GetCreateFlags()

	expectedURL := "/foo/bar/boot2docker.iso"
	os.Setenv("WRONG_BOOT2DOCKER_URL", expectedURL)
	defer os.Unsetenv("WRONG_BOOT2DOCKER_URL")

	driverOptions, err := prepareDriverOptions(supportedFlags, make(map[string]interface{}))
	assert.NoError(t, err, "Error preparing driver options")

	err = d.SetConfigFromFlags(driverOptions)
	assert.NoError(t, err, "Error setting configs from flags")

	assert.Empty(t, d.Boot2DockerURL)
}

func Test_unused_explicit_driver_options_returns_error(t *testing.T) {
	d := &tests.MockDriver{
		BaseDriver: drivers.BaseDriver{},
	}
	supportedFlags := d.GetCreateFlags()

	explicitConfig := make(map[string]interface{})
	explicitConfig["foo"] = "bar"

	_, err := prepareDriverOptions(supportedFlags, explicitConfig)
	assert.Error(t, err, "preparingDriverOptions should return error")

	expectedError := "Unused explicit driver options: map[foo:bar]"
	assert.EqualError(t, err, expectedError)
}
