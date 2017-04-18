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

package oc

import (
	"fmt"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util"

	"github.com/minishift/minishift/pkg/util/cmd"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/pkg/errors"
	"io"
	"strings"
)

const (
	invalidOcPathError         = "The specified path to oc '%s' does not exist"
	invalidKubeConfigPathError = "The specified path to the kube config '%s' does not exist"
)

type OcRunner struct {
	ocPath         string
	kubeConfigPath string
	runner         util.Runner
}

// NewOcRunner creates a new OcRunner which uses the oc binary specified via the ocPath parameter. An error is returned
// in case the oc binary does not exist or is not executable.
func NewOcRunner(ocPath string, kubeConfigPath string) (*OcRunner, error) {
	if !filehelper.Exists(ocPath) {
		return nil, errors.New(fmt.Sprintf(invalidOcPathError, ocPath))
	}

	if !filehelper.Exists(kubeConfigPath) {
		return nil, errors.New(fmt.Sprintf(invalidKubeConfigPathError, kubeConfigPath))
	}

	return &OcRunner{ocPath: ocPath, kubeConfigPath: kubeConfigPath, runner: util.RealRunner{}}, nil
}

func (oc *OcRunner) Run(command string, stdOut io.Writer, stdErr io.Writer) int {
	args := cmd.SplitCmdString(command)

	// make sure we run with our copy of kube config to not influence the user
	args = append([]string{fmt.Sprintf("--config=%s", oc.kubeConfigPath)}, args...)

	return oc.runner.Run(stdOut, stdErr, oc.ocPath, args...)
}

func (oc *OcRunner) RunAsUser(command string, stdOut io.Writer, stdErr io.Writer) int {
	args := strings.Split(command, " ")
	return oc.runner.Run(stdOut, stdErr, oc.ocPath, args...)
}

// AddSudoerRoleForUser gives the specified user the sudoer role
// See also https://docs.openshift.org/latest/architecture/additional_concepts/authentication.html#authentication-impersonation
func (oc *OcRunner) AddSudoerRoleForUser(user string) error {
	cmd := fmt.Sprintf("adm policy add-cluster-role-to-user sudoer %s", user)
	exitCode := oc.Run(cmd, nil, nil)
	if exitCode != 0 {
		return errors.New("Unable to add sudoer role")
	}

	return nil
}

// AddCliContext adds a CLI context for the user and namespace for the current OpenShift cluster. See also
// https://docs.openshift.com/enterprise/3.0/cli_reference/manage_cli_profiles.html
func (oc *OcRunner) AddCliContext(context string, ip string, username string, namespace string) error {
	ip = strings.Replace(ip, ".", "-", -1)
	cmd := fmt.Sprintf("config set-context %s --cluster=%s:%d --user=%s/%s:%d --namespace=%s", context, ip, constants.APIServerPort, username, ip, constants.APIServerPort, namespace)

	exitCode := oc.RunAsUser(cmd, nil, nil)
	if exitCode != 0 {
		return errors.New("Unable to create CLI context")
	}

	cmd = fmt.Sprintf("config use-context %s", context)
	exitCode = oc.RunAsUser(cmd, nil, nil)
	if exitCode != 0 {
		return errors.New("Unable to switch CLI context")
	}

	return nil
}
