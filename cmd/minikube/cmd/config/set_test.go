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
	"testing"
	"io/ioutil"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"os"
	"path/filepath"
)

func TestNotFound(t *testing.T) {
	err := set("nonexistant", "10")
	if err == nil {
		t.Fatalf("Set did not return error for unknown property")
	}
}

func TestModifyData(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-config-")
	if err != nil {
		t.Error()
	}
	constants.ConfigFile = filepath.Join(testDir, "config.json")
	defer os.RemoveAll(testDir)
	err = set("cpus", "4")
	if err != nil {
		t.Fatalf("Error setting value %s", err)
	}
	// override existing value and check if that persistent
	err = set("cpus", "10")
	if err != nil {
		t.Fatalf("Error setting value %s", err)
	}
	getValue, err := get("cpus")
	if err != nil {
		t.Fatalf("Error getting value %s", err)
	}
	if getValue != "10" {
		t.Fatal("Not able to update data to config file")
	}
}
