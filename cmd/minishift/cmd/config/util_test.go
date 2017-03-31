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

import "testing"

var minikubeConfig = MinishiftConfig{
	"vm-driver":            "kvm",
	"cpus":                 12,
	"show-libmachine-logs": true,
}

func TestFindSettingNotFound(t *testing.T) {
	s, err := findSetting("nonexistant")
	if err == nil {
		t.Fatalf("Unexpected setting found. [%+v]", s)
	}
}

func TestFindSetting(t *testing.T) {
	s, err := findSetting("vm-driver")
	if err != nil {
		t.Fatalf("Cannot find the setting of the vm-driver: %s", err)
	}
	if s.name != "vm-driver" {
		t.Fatalf("Incorrect setting, expected vm-driver, received %s", s.name)
	}
}

func TestSetString(t *testing.T) {
	err := SetString(minikubeConfig, "vm-driver", "virtualbox")
	if err != nil {
		t.Fatalf("Cannot set the string: %s", err)
	}
}

func TestSetInt(t *testing.T) {
	err := SetInt(minikubeConfig, "cpus", "22")
	if err != nil {
		t.Fatalf("Cannot set integer value in config: %s", err)
	}
	val, ok := minikubeConfig["cpus"].(int)
	if !ok {
		t.Fatalf("Type is not set to int")
	}
	if val != 22 {
		t.Fatalf("SetInt set value is incorrect")
	}
}

func TestSetBool(t *testing.T) {
	err := SetBool(minikubeConfig, "show-libmachine-logs", "true")
	if err != nil {
		t.Fatalf("Cannot set boolean value in config: %s", err)
	}
	val, ok := minikubeConfig["show-libmachine-logs"].(bool)
	if !ok {
		t.Fatalf("Type is not set to bool")
	}
	if !val {
		t.Fatalf("SetBool set value is incorrect")
	}
}
