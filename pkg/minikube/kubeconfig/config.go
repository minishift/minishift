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
	"errors"
	"fmt"
	"github.com/minishift/minishift/pkg/util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

// kubeconfig Data types
type ClusterType struct {
	Cluster map[string]string `yaml:"cluster"`
	Name    string            `yaml:"name"`
}

type ContextType struct {
	Context map[string]string `yaml:"context"`
	Name    string            `yaml:"name"`
}

type UserType struct {
	User map[string]string `yaml:"user"`
	Name string            `yaml:"name"`
}

type SystemKubeConfig struct {
	ApiVersion     string        `yaml:"apiVersion"`
	Clusters       []ClusterType `yaml:"clusters"`
	Contexts       []ContextType `yaml:"contexts"`
	CurrentContext string        `yaml:"current-context"`
	Users          []UserType    `yaml:"users"`
}

func GetConfigPath() string {
	var configPath = filepath.Join(util.HomeDir(), ".kube", "config")
	if os.Getenv("KUBECONFIG") != "" {
		configPath = os.Getenv("KUBECONFIG")
	}
	return configPath
}

// Cache system admin entries to be used to run oc commands
func CacheSystemAdminEntries(systemEntriesConfigPath, clusterName string) error {
	config, err := Read(GetConfigPath())
	if err != nil {
		return errors.New(fmt.Sprintf("Error reading config file %s", systemEntriesConfigPath))
	}

	targetConfig := SystemKubeConfig{ApiVersion: "v1"}
	for k, v := range config.Clusters {
		if v.Name == clusterName {
			targetConfig.Clusters = append(targetConfig.Clusters, config.Clusters[k])
			break
		}
	}

	targetConfig.CurrentContext = fmt.Sprintf("default/%s/system:admin", clusterName)
	for k, v := range config.Contexts {
		if v.Name == targetConfig.CurrentContext {
			targetConfig.Contexts = append(targetConfig.Contexts, config.Contexts[k])
			break
		}
	}

	userName := fmt.Sprintf("system:admin/%s", clusterName)
	for k, v := range config.Users {
		if v.Name == userName {
			targetConfig.Users = append(targetConfig.Users, config.Users[k])
			break
		}
	}

	yamlData, err := yaml.Marshal(&targetConfig)
	if err != nil {
		return errors.New("Error marshalling system kubeconfig entries")
	}
	// Write to machines/<MACHINE_NAME>_kubeconfig
	if err = ioutil.WriteFile(systemEntriesConfigPath, yamlData, 0644); err != nil {
		return err
	}

	return nil
}

func Read(configPath string) (SystemKubeConfig, error) {
	config := SystemKubeConfig{}
	data, _ := ioutil.ReadFile(configPath)
	err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
