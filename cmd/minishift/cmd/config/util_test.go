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

package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var minikubeConfig = MinishiftConfig{
	"vm-driver":            "kvm",
	"cpus":                 12,
	"show-libmachine-logs": true,
}

func TestFindSettingNotFound(t *testing.T) {
	_, err := findSetting("nonexistant")
	assert.Error(t, err)
}

func TestFindSetting(t *testing.T) {
	s, err := findSetting("vm-driver")
	assert.NoError(t, err, "Cannot find the setting of the vm-driver")
	assert.Equal(t, "vm-driver", s.Name, "Setting of the vm-driver not found")
}

func TestSetString(t *testing.T) {
	err := SetString(minikubeConfig, "vm-driver", "virtualbox")
	assert.NoError(t, err, "Error setting config")
}

func TestSetInt(t *testing.T) {
	err := SetInt(minikubeConfig, "cpus", "22")

	assert.NoError(t, err, "Error setting int value")

	val, ok := minikubeConfig["cpus"].(int)

	assert.True(t, ok, "Type is not set to int")
	assert.Equal(t, 22, val, "SetInt set value is incorrect")
	assert.IsType(t, *new(int), minikubeConfig["cpus"])
}

func TestSetBool(t *testing.T) {
	err := SetBool(minikubeConfig, "show-libmachine-logs", "true")

	assert.NoError(t, err, "Error setting boolean value in config")

	val, ok := minikubeConfig["show-libmachine-logs"].(bool)

	assert.True(t, ok, "Type is not set to bool")
	assert.True(t, val, "SetBool set value is incorrect")
	assert.IsType(t, *new(bool), minikubeConfig["show-libmachine-logs"])
}

func TestSetSlice(t *testing.T) {
	expectedSlice := []string{"172.0.0.1/16", "mycustom.registry.com/3030"}
	err := SetSlice(minikubeConfig, "insecure-registry", strings.Join(expectedSlice, ","))
	assert.NoError(t, err, "Error setting slice")
	val, _ := minikubeConfig["insecure-registry"].([]string)
	assert.IsType(t, *new([]string), minikubeConfig["insecure-registry"])
	assert.Equal(t, expectedSlice, val)
}
