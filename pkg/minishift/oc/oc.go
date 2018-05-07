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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/minishift/minishift/pkg/minikube/constants"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	openshiftVersionCheck "github.com/minishift/minishift/pkg/minishift/openshift/version"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/cmd"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/pkg/errors"
)

const (
	invalidOcPathError         = "The specified path to oc '%s' does not exist"
	invalidKubeConfigPathError = "The specified path to the kube config '%s' does not exist"
)

type OcRunner struct {
	OcPath         string
	KubeConfigPath string
	Runner         util.Runner
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

	return &OcRunner{OcPath: ocPath, KubeConfigPath: kubeConfigPath, Runner: util.RealRunner{}}, nil
}

func (oc *OcRunner) Run(command string, stdOut io.Writer, stdErr io.Writer) int {
	args := cmd.SplitCmdString(command)

	// make sure we run with our copy of kube config to not influence the user
	args = append([]string{fmt.Sprintf("--config=%s", oc.KubeConfigPath)}, args...)

	return oc.Runner.Run(stdOut, stdErr, oc.OcPath, args...)
}

func (oc *OcRunner) RunAsUser(command string, stdOut io.Writer, stdErr io.Writer) int {
	args := strings.Split(command, " ")
	return oc.Runner.Run(stdOut, stdErr, oc.OcPath, args...)
}

// AddSudoerRoleForUser gives the specified user the sudoer role
// See also https://docs.openshift.org/latest/architecture/additional_concepts/authentication.html#authentication-impersonation
func (oc *OcRunner) AddSudoerRoleForUser(user string) error {
	cmd := fmt.Sprintf("adm policy add-cluster-role-to-user sudoer %s", user)
	errorBuffer := new(bytes.Buffer)
	exitCode := oc.Run(cmd, nil, errorBuffer)
	if exitCode != 0 {
		return fmt.Errorf("Unable to add sudoer role: %v", errorBuffer)
	}

	return nil
}

// AddSystemAdminEntrytoKubeConfig adds the system:admin certs to ~/.kube/config
// This function actually mimic the KUBECONFIG magical power of merging different kubeconfig file.
// KUBECONFIG=~config1:config2 oc config view  --flatten
// https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/#set-the-kubeconfig-environment-variable
func (oc *OcRunner) AddSystemAdminEntryToKubeConfig(ocPath string) error {
	user, err := user.Current()
	if err != nil {
		return fmt.Errorf("unable to find current user: %v", err)
	}

	minishiftKubeConfigPath := oc.KubeConfigPath
	kubeConfigGlobalPath := filepath.Join(user.HomeDir, ".kube", "config")

	// For Linux and Mac, the list is colon-delimited. For Windows, the list is semicolon-delimited for kubeconfig  env variable
	// https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#the-kubeconfig-environment-variable
	kubeconfig := fmt.Sprintf("%s:%s", minishiftKubeConfigPath, kubeConfigGlobalPath)
	if runtime.GOOS == "windows" {
		kubeconfig = fmt.Sprintf("%s;%s", minishiftKubeConfigPath, kubeConfigGlobalPath)
	}
	realrunner := util.RealRunner{Env: []string{fmt.Sprintf("KUBECONFIG=%s", kubeconfig)}}

	outBuffer := new(bytes.Buffer)
	cmdArgs := []string{"config", "view", "--flatten"}
	exitCode := realrunner.Run(outBuffer, nil, ocPath, cmdArgs...)
	if exitCode != 0 {
		return fmt.Errorf("unable to view kubeconfig file")
	}

	if err := os.MkdirAll(filepath.Join(user.HomeDir, ".kube"), 0755); err != nil {
		return fmt.Errorf("unable to create kubeconfig dir ($HOME/.kube): %v", err)
	}

	if err := ioutil.WriteFile(kubeConfigGlobalPath, outBuffer.Bytes(), 0600); err != nil {
		return fmt.Errorf("unable to write in kubeconfig file %s: %v", kubeConfigGlobalPath, err)
	}
	return nil
}

// AddCliContext adds a CLI context for the user and namespace for the current OpenShift cluster. See also
// https://docs.openshift.com/enterprise/3.0/cli_reference/manage_cli_profiles.html
func (oc *OcRunner) AddCliContext(context string, ip string, username string, namespace string, runner util.Runner, ocPath string) error {
	cmdArgs := []string{"login",
		fmt.Sprintf("-u=%s", username),
		fmt.Sprintf("-p=%s", minishiftConstants.DefaultUserPassword),
		fmt.Sprintf("%s:8443", ip)}

	stdBuffer := new(bytes.Buffer)
	exitCode := runner.Run(stdBuffer, os.Stderr, ocPath, cmdArgs...)
	if exitCode != 0 {
		return fmt.Errorf("Unable to login to cluster")
	}

	ip = strings.Replace(ip, ".", "-", -1)
	cmd := fmt.Sprintf("config set-context %s --cluster=%s:%d --user=%s/%s:%d --namespace=%s", context, ip, constants.APIServerPort, username, ip, constants.APIServerPort, namespace)
	errorBuffer := new(bytes.Buffer)
	exitCode = oc.RunAsUser(cmd, nil, errorBuffer)
	if exitCode != 0 {
		return fmt.Errorf("Unable to create CLI context: %v", errorBuffer)
	}

	cmd = fmt.Sprintf("config use-context %s", context)
	exitCode = oc.RunAsUser(cmd, nil, errorBuffer)
	if exitCode != 0 {
		return fmt.Errorf("Unable to switch CLI context: %v", errorBuffer)
	}

	return nil
}

func SupportFlag(flag string, ocPath string, runner util.Runner) bool {
	var buffer bytes.Buffer
	cmdArgs := []string{"cluster", "up", "-h"}
	runner.Run(&buffer, os.Stderr, ocPath, cmdArgs...)
	ocCommandOptions := parseOcHelpCommand(buffer.Bytes())
	if ocCommandOptions != nil {
		return flagExist(ocCommandOptions, flag)
	}
	return false
}

func parseOcHelpCommand(cmdOut []byte) []string {
	ocOptions := []string{}
	var openshiftVersion string
	ocOptionRegex := regexp.MustCompile(`(?s)Options(.*)OpenShift images`)
	if instanceState.InstanceStateConfig != nil {
		openshiftVersion = instanceState.InstanceStateConfig.OpenshiftVersion
	}
	valid, _ := openshiftVersionCheck.IsGreaterOrEqualToBaseVersion(openshiftVersion, constants.RefactoredOcVersion)
	if valid {
		ocOptionRegex = regexp.MustCompile(`(?s)Options(.*)host config dir`)
	}
	matches := ocOptionRegex.FindSubmatch(cmdOut)
	if matches != nil {
		tmpOptionsList := string(matches[0])
		for _, value := range strings.Split(tmpOptionsList, "\n")[1:] {
			tmpOption := strings.Split(strings.Split(strings.TrimSpace(value), "=")[0], "--")
			if len(tmpOption) > 1 {
				ocOptions = append(ocOptions, tmpOption[1])
			}
		}
	} else {
		return nil
	}
	return ocOptions
}

func flagExist(ocCommandOptions []string, flag string) bool {
	for _, v := range ocCommandOptions {
		if v == flag {
			return true
		}
	}
	return false
}
