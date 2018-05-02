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

package openshift

import (
	"errors"
	"fmt"
	"strings"

	"github.com/minishift/minishift/pkg/minikube/constants"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
)

func GetDockerRegistryInfo(registryAddonEnabled bool, openshiftVersion string) (string, error) {
	var registryInfo string
	var err error
	if isRegistryRouteEnabled(registryAddonEnabled) {
		registryInfo, err = fetchRegistryRoute(openshiftVersion)
	} else {
		registryInfo, err = fetchRegistryService()
	}
	return registryInfo, err
}

func isRegistryRouteEnabled(registryAddonEnabled bool) bool {
	if !registryAddonEnabled {
		return false
	}
	namespace := "default"
	cmdArgText := fmt.Sprintf("get route docker-registry -n %s --config=%s", namespace, constants.KubeConfigPath)
	tokens := strings.Split(cmdArgText, " ")
	cmdName := instanceState.InstanceConfig.OcPath
	_, err := runner.Output(cmdName, tokens...)
	if err != nil {
		return false
	}
	return true
}

func fetchRegistryRoute(openshiftVersion string) (string, error) {
	namespace := "default"
	route := "route/docker-registry"
	cmdArgText := fmt.Sprintf("get -o jsonpath={.spec.host} %s -n %s --config=%s", route, namespace, constants.KubeConfigPath)
	tokens := strings.Split(cmdArgText, " ")
	cmdName := instanceState.InstanceConfig.OcPath
	cmdOut, err := runner.Output(cmdName, tokens...)
	if err != nil {
		return "", err
	}

	registryInfo := string(cmdOut)
	return registryInfo, nil
}

func fetchRegistryService() (string, error) {
	namespace := "default"
	service := "service/docker-registry"
	cmdArgText := fmt.Sprintf("get -o jsonpath={.spec.clusterIP}:{.spec.ports[*].port} %s -n %s --config=%s", service, namespace, constants.KubeConfigPath)
	tokens := strings.Split(cmdArgText, " ")
	cmdName := instanceState.InstanceConfig.OcPath
	cmdOut, err := runner.Output(cmdName, tokens...)
	if err != nil {
		return "", errors.New(fmt.Sprintf("No information found for '%s'", service))
	}
	registryInfo := string(cmdOut)
	return registryInfo, nil
}
