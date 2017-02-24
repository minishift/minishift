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

package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var testDir string

func TestNewInstanceConfig(t *testing.T) {
	setup(t)
	defer teardown()

	expectedFilePath := filepath.Join(testDir, "fake-machine.json")
	cfg, _ := NewInstanceConfig(expectedFilePath)

	if cfg.FilePath != expectedFilePath {
		t.Errorf("Expected path '%s'. Received '%s'", expectedFilePath, cfg.FilePath)
	}

	if _, err := os.Stat(cfg.FilePath); os.IsNotExist(err) {
		t.Errorf("File %s should exists", cfg.FilePath)
	}
}

func TestConfigOnFileExists(t *testing.T) {
	setup(t)
	defer teardown()

	filePath := filepath.Join(testDir, "fake-machine.json")
	expectedOcPath := filepath.Join(testDir, "fakeOc")
	cfg := &InstanceConfigType{
		FilePath: filePath,
		OcPath:   expectedOcPath,
	}

	jsonData, _ := json.MarshalIndent(cfg, "", "\t")
	// create config file before NewInstanceConfig
	ioutil.WriteFile(cfg.FilePath, jsonData, 0644)

	newCfg, _ := NewInstanceConfig(filePath)
	if newCfg.OcPath != expectedOcPath {
		t.Errorf("Expected oc path '%s'. Received '%s'", expectedOcPath, cfg.OcPath)
	}
}

func TestWrite(t *testing.T) {
	setup(t)
	defer teardown()

	path := filepath.Join(testDir, "fake-machine.json")
	cfg, _ := NewInstanceConfig(path)

	expectedOcPath := filepath.Join(testDir, "fakeOc")
	cfg.OcPath = expectedOcPath
	cfg.Write()

	// read config file and verify content
	var testCfg *InstanceConfigType
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		t.Errorf("Error reading config file %s", path)
	}

	json.Unmarshal(raw, &testCfg)
	if testCfg.OcPath != cfg.OcPath {
		t.Errorf("Expected oc path '%s'. Received '%s'", expectedOcPath, cfg.OcPath)
	}
}

func TestDelete(t *testing.T) {
	setup(t)
	defer teardown()

	path := filepath.Join(testDir, "fake-machine.json")
	cfg, _ := NewInstanceConfig(path)

	cfg.Delete()

	if _, err := os.Stat(cfg.FilePath); err == nil {
		t.Errorf("Expected file '%s' to be deleted", path)
	}
}

func setup(t *testing.T) {
	var err error
	testDir, err = ioutil.TempDir("", "minishift-test-config-")

	if err != nil {
		t.Error(err)
	}
}

// teardown remove the temp directory
func teardown() {
	os.RemoveAll(testDir)
}
