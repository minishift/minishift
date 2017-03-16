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

package kubeconfig

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var (
	testKubeConfigPath   string
	testSystemConfigPath string
)

func TestCacheSystemAdminEntries(t *testing.T) {
	setUp()
	defer tearDown(t)
	testSystemConfig := SystemKubeConfig{}

	CacheSystemAdminEntries(testSystemConfigPath, "10-168-99-100:8443")

	// read test kube config
	data, _ := ioutil.ReadFile(testKubeConfigPath)
	kubeData := string(data)

	// read test system config
	data, _ = ioutil.ReadFile(testSystemConfigPath)
	yaml.Unmarshal([]byte(data), &testSystemConfig)

	// verify entries
	strings.Contains(kubeData, testSystemConfig.Users[0].Name)
	strings.Contains(kubeData, testSystemConfig.Users[0].User["client-certificate-data"])
	strings.Contains(kubeData, testSystemConfig.Clusters[0].Name)
	strings.Contains(kubeData, testSystemConfig.Clusters[0].Cluster["certificate-authority-data"])
	strings.Contains(kubeData, testSystemConfig.Clusters[0].Cluster["server"])
	strings.Contains(kubeData, testSystemConfig.Contexts[0].Name)
	strings.Contains(kubeData, testSystemConfig.Contexts[0].Context["cluster"])
	strings.Contains(kubeData, testSystemConfig.Contexts[0].Context["user"])
}

func setUp() {
	workingDir, _ := os.Getwd()
	testKubeConfigPath = fmt.Sprintf("%s/test_kube_config", workingDir)

	os.Setenv("KUBECONFIG", testKubeConfigPath)
	testSystemConfigPath = workingDir + "/test_systemconfig"
}

func tearDown(t *testing.T) {
	if err := os.Remove(testSystemConfigPath); err != nil {
		t.Fatalf("Error delete file %s", testSystemConfigPath)
	}
}
