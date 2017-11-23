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

	"github.com/stretchr/testify/assert"
)

var testDir string

func TestNewInstanceConfig(t *testing.T) {
	setup(t)
	defer teardown()

	expectedFilePath := filepath.Join(testDir, "fake-machine.json")
	cfg, _ := NewInstanceConfig(expectedFilePath)
	assert.Equal(t, expectedFilePath, cfg.FilePath)

	_, err := os.Stat(cfg.FilePath)
	assert.NoError(t, err, "File %s should exists", cfg.FilePath)
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
	assert.Equal(t, expectedOcPath, newCfg.OcPath)
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
	assert.NoError(t, err, "Error in reading config file %s", path)

	json.Unmarshal(raw, &testCfg)
	assert.Equal(t, testCfg.OcPath, cfg.OcPath)
}

func TestDelete(t *testing.T) {
	setup(t)
	defer teardown()

	path := filepath.Join(testDir, "fake-machine.json")
	cfg, _ := NewInstanceConfig(path)

	cfg.Delete()

	_, err := os.Stat(cfg.FilePath)
	assert.Error(t, err)
}

func setup(t *testing.T) {
	var err error
	testDir, err = ioutil.TempDir("", "minishift-test-config-")

	assert.NoError(t, err, "Error creating temp directory")
}

// teardown remove the temp directory
func teardown() {
	os.RemoveAll(testDir)
}
