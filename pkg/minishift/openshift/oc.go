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
	"encoding/json"
	"errors"
	"fmt"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util"
	routeApi "github.com/openshift/origin/pkg/route/api"
	"strings"
)

// runner executes commands on the host
var runner util.Runner = &util.RealRunner{}

// Add developer user to cluster sudoers
func AddSudoersRoleForUser(user string) error {
	cmdName := instanceState.Config.OcPath
	cmdArgs := []string{"login", "-u", "system:admin"}
	if _, err := runner.Output(cmdName, cmdArgs...); err != nil {
		return err
	}
	// https://docs.openshift.org/latest/architecture/additional_concepts/authentication.html#authentication-impersonation
	cmdArgs = []string{"adm", "policy", "add-cluster-role-to-user", "sudoer", user}
	if _, err := runner.Output(cmdName, cmdArgs...); err != nil {
		return err
	}
	cmdArgs = []string{"login", "-u", user}
	if _, err := runner.Output(cmdName, cmdArgs...); err != nil {
		return err
	}
	return nil
}

// Get the route for service
func GetServiceURL(service, namespace string, https bool) (string, error) {
	urlScheme := "http://"
	if https {
		urlScheme = "https://"
	}

	cmdName := instanceState.Config.OcPath
	cmdArgText := fmt.Sprintf("get route/%s", service)
	if !isProjectExists(namespace) {
		return "", errors.New(fmt.Sprintf("Namespace %s doesn't exits", namespace))
	}
	if namespace != "" {
		cmdArgText = cmdArgText + " -n " + namespace
	}

	// json filter
	cmdArgText += " -o json"
	cmdOut, err := runner.Output(cmdName, strings.Split(cmdArgText, " ")...)
	if err != nil {
		return "", err
	}

	var route routeApi.Route
	json.Unmarshal(cmdOut, &route)

	return urlScheme + route.Spec.Host, nil
}

// Check whether project exists or not
func isProjectExists(project string) bool {
	if project == "" {
		return true
	}

	cmdName := instanceState.Config.OcPath
	cmdArgs := []string{"get", "projects", project}
	_, err := runner.Output(cmdName, cmdArgs...)
	if err != nil {
		return false
	}
	return true
}
