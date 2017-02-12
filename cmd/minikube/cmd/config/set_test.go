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
	"github.com/minishift/minishift/pkg/minikube/constants"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNotFound(t *testing.T) {
	err := set("nonexistant", "10")
	if err == nil {
		t.Fatal("Set did not return error for unknown property")
	}
}

func TestModifyData(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-config-")
	if err != nil {
		t.Error()
	}
	constants.ConfigFile = filepath.Join(testDir, "config.json")
	defer os.RemoveAll(testDir)

	verifyValueUnset(t, "cpus")

	persistValue(t, "cpus", "4")
	verifyStoredValue(t, "cpus", "4")

	// override existing value and check if that persistent
	persistValue(t, "cpus", "10")
	verifyStoredValue(t, "cpus", "10")
}

func verifyStoredValue(t *testing.T, key string, expectedValue string) {
	actualValue, err := get(key)
	if err != nil {
		t.Fatalf("Error getting value %s", err)
	}
	if actualValue != expectedValue {
		t.Fatalf("Unexpexted value in confif. Expected '%s'. Got '%s'.", expectedValue, actualValue)
	}
}

func verifyValueUnset(t *testing.T, key string) {
	actualValue, err := get(key)
	if err != nil {
		t.Fatalf("Error getting value %s", err)
	}
	if actualValue != "<nil>" {
		t.Fatalf("Expexted '<nil>' value. Got '%s'.", actualValue)
	}
}

func persistValue(t *testing.T, key string, value string) {
	err := set(key, value)
	if err != nil {
		t.Fatalf("Error setting value %s", err)
	}
}
