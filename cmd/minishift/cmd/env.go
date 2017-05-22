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

// Part of this code is heavily inspired/copied by the following file:
// github.com/docker/machine/commands/env.go

package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/minishift/minishift/pkg/minikube/constants"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/minishift/minishift/pkg/util/shell"
	"github.com/spf13/cobra"
	"strings"
)

const (
	envTmpl = `{{ .Prefix }}DOCKER_TLS_VERIFY{{ .Delimiter }}{{ .DockerTLSVerify }}{{ .Suffix }}{{ .Prefix }}DOCKER_HOST{{ .Delimiter }}{{ .DockerHost }}{{ .Suffix }}{{ .Prefix }}DOCKER_CERT_PATH{{ .Delimiter }}{{ .DockerCertPath }}{{ .Suffix }}{{ .Prefix }}DOCKER_API_VERSION{{ .Delimiter }}{{ .DockerAPIVersion }}{{ .Suffix }}{{ if .NoProxyVar }}{{ .Prefix }}{{ .NoProxyVar }}{{ .Delimiter }}{{ .NoProxyValue }}{{ .Suffix }}{{end}}{{ .UsageHint }}`
)

var (
	noProxy    bool
	forceShell string
	unset      bool
)

type DockerShellConfig struct {
	shell.ShellConfig
	DockerCertPath   string
	DockerHost       string
	DockerTLSVerify  string
	DockerAPIVersion string
	UsageHint        string
	NoProxyVar       string
	NoProxyValue     string
}

func getConfigSet(api libmachine.API, forceShell string, noProxy bool) (*DockerShellConfig, error) {
	envMap, err := cluster.GetHostDockerEnv(api)
	if err != nil {
		return nil, err
	}

	userShell, err := shell.GetShell(forceShell)
	if err != nil {
		return nil, err
	}

	cmdLine := "minishift docker-env"
	shellCfg := &DockerShellConfig{
		DockerCertPath:   envMap["DOCKER_CERT_PATH"],
		DockerHost:       envMap["DOCKER_HOST"],
		DockerTLSVerify:  envMap["DOCKER_TLS_VERIFY"],
		DockerAPIVersion: envMap["DOCKER_API_VERSION"],
		UsageHint:        shell.GenerateUsageHint(userShell, cmdLine),
	}

	if noProxy {
		host, err := api.Load(constants.MachineName)
		if err != nil {
			return nil, fmt.Errorf("Error getting IP: %s", err)
		}

		ip, err := host.Driver.GetIP()
		if err != nil {
			return nil, fmt.Errorf("Error getting host IP: %s", err)
		}

		noProxyVar, noProxyValue := shell.FindNoProxyFromEnv()

		// add the docker host to the no_proxy list idempotently
		switch {
		case noProxyValue == "":
			noProxyValue = ip
		case strings.Contains(noProxyValue, ip):
		//ip already in no_proxy list, nothing to do
		default:
			noProxyValue = fmt.Sprintf("%s,%s", noProxyValue, ip)
		}

		shellCfg.NoProxyVar = noProxyVar
		shellCfg.NoProxyValue = noProxyValue
	}

	shellCfg.Prefix, shellCfg.Suffix, shellCfg.Delimiter = shell.GetPrefixSuffixDelimiterForSet(userShell, false)

	return shellCfg, nil
}

func getConfigUnset(forceShell string, noProxy bool) (*DockerShellConfig, error) {
	userShell, err := shell.GetShell(forceShell)
	if err != nil {
		return nil, err
	}

	cmdLine := "minishift docker-env"
	shellCfg := &DockerShellConfig{
		UsageHint: shell.GenerateUsageHint(userShell, cmdLine),
	}

	if noProxy {
		shellCfg.NoProxyVar, shellCfg.NoProxyValue = shell.FindNoProxyFromEnv()
	}

	prefix, suffix, delimiter := shell.GetPrefixSuffixDelimiterForUnSet(userShell)
	shellCfg.Prefix = prefix
	shellCfg.Suffix = suffix
	shellCfg.Delimiter = delimiter

	return shellCfg, nil
}

func executeTemplateStdout(shellCfg *DockerShellConfig) error {
	tmpl := template.Must(template.New("envConfig").Parse(envTmpl))
	return tmpl.Execute(os.Stdout, shellCfg)
}

// envCmd represents the docker-env command
var dockerEnvCmd = &cobra.Command{
	Use:   "docker-env",
	Short: "Sets Docker environment variables.",
	Long:  `Sets Docker environment variables, similar to '$(docker-machine env)'.`,
	Run: func(cmd *cobra.Command, args []string) {

		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		util.ExitIfUndefined(api, constants.MachineName)

		host, err := api.Load(constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		util.ExitIfNotRunning(host.Driver)

		var shellCfg *DockerShellConfig
		if unset {
			shellCfg, err = getConfigUnset(forceShell, noProxy)
			if err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Error unsetting environment variables: %s", err.Error()))
			}
		} else {
			shellCfg, err = getConfigSet(api, forceShell, noProxy)
			if err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Error setting environment variables: %s", err.Error()))
			}
		}

		executeTemplateStdout(shellCfg)
	},
}

func init() {
	RootCmd.AddCommand(dockerEnvCmd)
	dockerEnvCmd.Flags().BoolVar(&noProxy, "no-proxy", false, "Add the virtual machine IP to the NO_PROXY environment variable.")
	dockerEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Force setting the environment for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
	dockerEnvCmd.Flags().BoolVarP(&unset, "unset", "u", false, "Clear the environment variable values instead of setting them.")
}
