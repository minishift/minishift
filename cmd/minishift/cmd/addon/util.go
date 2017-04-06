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

package addon

import (
	"fmt"
	"reflect"

	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/addon"
	"github.com/minishift/minishift/pkg/minishift/addon/command"
	"github.com/minishift/minishift/pkg/minishift/addon/manager"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

const (
	ip_key             = "ip"
	routing_suffix_key = "routing-suffix"
)

// getAddOnManager returns the addon manager
func GetAddOnManager() *manager.AddOnManager {
	addOnConfigs := getAddOnConfiguration()
	m, err := manager.NewAddOnManager(constants.MakeMiniPath("addons"), addOnConfigs)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Unable to initialize the add-on manager: %s", err.Error()))
	}

	return m
}

func GetExecutionContext(ip string, routingSuffix string, ocPath string, kubeConfigPath string, sshCommander provision.SSHCommander) *command.ExecutionContext {
	context, err := command.NewExecutionContext(ocPath, kubeConfigPath, sshCommander)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Unable to initialise execution context: %s", err.Error()))
	}

	context.AddToContext(ip_key, ip)
	context.AddToContext(routing_suffix_key, routingSuffix)

	return context
}

func writeAddOnConfig(addOnConfigMap map[string]*addon.AddOnConfig) {
	c, err := config.ReadConfig()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Unable to read the Minishift configuration: %s", err.Error()))
	}

	c[addOnConfigKey] = addOnConfigMap

	err = config.WriteConfig(c)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Unable to write the Minishift configuration: %s", err.Error()))
	}
}

// getAddOnConfiguration reads the Minishift configuration in $MINISHIFT_HOME/config/config.json related to addons and returns
// a map of addon names to AddOnConfig
func getAddOnConfiguration() map[string]*addon.AddOnConfig {
	c, err := config.ReadConfig()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Unable to read the Minishift configuration: %s", err.Error()))
	}

	var configSlice map[string]interface{}
	if c[addOnConfigKey] != nil {
		configSlice = c[addOnConfigKey].(map[string]interface{})
	} else {
		configSlice = make(map[string]interface{})
	}

	addOnConfigs := make(map[string]*addon.AddOnConfig)
	for _, entry := range configSlice {
		addOnConfig := &addon.AddOnConfig{}
		addOnMap := entry.(map[string]interface{})
		fillStruct(addOnMap, addOnConfig)
		addOnConfigs[addOnConfig.Name] = addOnConfig
	}

	return addOnConfigs
}

// fillStruct populates the specified result struct with the data provided in the data map
func fillStruct(data map[string]interface{}, result interface{}) {
	t := reflect.ValueOf(result).Elem()
	for k, v := range data {
		val := t.FieldByName(k)
		val.Set(reflect.ValueOf(v))
	}
}
