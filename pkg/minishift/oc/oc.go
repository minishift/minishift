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
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/cmd"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"runtime"
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
// See also https://docs.okd.io/latest/architecture/additional_concepts/authentication.html#authentication-impersonation
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
func (oc *OcRunner) AddSystemAdminEntryToKubeConfig(ocPath string) error {
	var existingKubeConfig, newKubeConfig *clientcmdapi.Config

	minishiftKubeConfigPath := oc.KubeConfigPath
	globalKubeConfigPath, err := GetGlobalKubeConfigPath()
	if err != nil {
		return err
	}
	if glog.V(2) {
		fmt.Println("Using Kubeconfig Path: ", globalKubeConfigPath)
	}
	dir, _ := filepath.Split(globalKubeConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("unable to create kubeconfig dir %s: %v", dir, err)
	}

	// Make sure .kube/config exist if not then this will create
	os.OpenFile(globalKubeConfigPath, os.O_RDONLY|os.O_CREATE, 0600)

	existingKubeConfig, err = clientcmd.LoadFromFile(globalKubeConfigPath)
	if err != nil {
		return fmt.Errorf("Not able to load %s: %s", globalKubeConfigPath, err)
	}

	newKubeConfig, err = clientcmd.LoadFromFile(minishiftKubeConfigPath)
	if err != nil {
		return fmt.Errorf("Not able to load %s: %s", minishiftKubeConfigPath, err)
	}

	merged := mergeKubeConfigs([]*clientcmdapi.Config{existingKubeConfig, newKubeConfig})

	return clientcmd.WriteToFile(*merged, globalKubeConfigPath)
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
	ocOptionRegex := regexp.MustCompile(`(?s)Options(.*)host config dir`)
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

func mergeKubeConfigs(configs []*clientcmdapi.Config) *clientcmdapi.Config {
	mergedConfig := clientcmdapi.NewConfig()
	for _, config := range configs {
		if config == nil {
			continue
		}
		// merge clusters
		for cName, c := range config.Clusters {
			mergedConfig.Clusters[cName] = c
		}
		// merge authinfos
		for aName, a := range config.AuthInfos {
			mergedConfig.AuthInfos[aName] = a
		}
		// merge contexts
		for ctxName, ctx := range config.Contexts {
			mergedConfig.Contexts[ctxName] = ctx
		}
		// merge extensions
		for extName, ext := range config.Extensions {
			mergedConfig.Extensions[extName] = ext
		}
	}
	return mergedConfig
}

// GetGlobalKubeConfigPath returns the path to the first entry in KUBECONFIG environment variable
// or if KUBECONFIG not set then $HOME/.kube/config
func GetGlobalKubeConfigPath() (string, error) {
	globalKubeConfigPath := os.Getenv("KUBECONFIG")
	globalKubeConfigPathList := strings.FieldsFunc(globalKubeConfigPath, splitKubeConfig)
	if len(globalKubeConfigPathList) > 0 {
		// Tools should write to last entry in the KUBECONFIG file instead first one.
		// oc cluster up also follow same.
		globalKubeConfigPath = strings.FieldsFunc(globalKubeConfigPath, splitKubeConfig)[len(globalKubeConfigPathList)-1]
	} else {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("unable to find current user: %v", err)
		}
		globalKubeConfigPath = filepath.Join(usr.HomeDir, ".kube", "config")
	}
	return globalKubeConfigPath, nil
}

func splitKubeConfig(r rune) bool {
	if runtime.GOOS == "windows" {
		return r == ';'
	}
	return r == ':'
}
