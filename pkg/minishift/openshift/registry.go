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

	instanceState "github.com/minishift/minishift/pkg/minishift/config"
)

func GetDockerRegistryInfo() (string, error) {
	namespace := "default"
	service := "service/docker-registry"
	cmdArgText := fmt.Sprintf("get -o jsonpath={.spec.clusterIP}:{.spec.ports[*].port} %s -n %s --config=%s", service, namespace, systemKubeConfigPath)
	tokens := strings.Split(cmdArgText, " ")
	cmdName := instanceState.InstanceConfig.OcPath
	cmdOut, err := runner.Output(cmdName, tokens...)
	if err != nil {
		return "", errors.New(fmt.Sprintf("No information found for '%s'", service))
	}

	registryInfo := byteArrayToString(cmdOut)
	return registryInfo, nil
}
