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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/config"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/minishift/minishift/pkg/util/shell"
	"github.com/spf13/cobra"
)

const (
	ocEnvTmpl = `{{ .Prefix }}PATH{{ .Delimiter }}{{ .OcDirPath }}{{ .PathSuffix }}{{ if .NoProxyVar }}{{ .Prefix }}{{ .NoProxyVar }}{{ .Delimiter }}{{ .NoProxyValue }}{{ .Suffix }}{{end}}{{ .UsageHint }}`
)

type OcShellConfig struct {
	shell.ShellConfig
	OcDirPath    string
	UsageHint    string
	NoProxyVar   string
	NoProxyValue string
}

func getOcShellConfig(api libmachine.API, ocPath string, forcedShell string, noProxy bool) (*OcShellConfig, error) {
	userShell, err := shell.GetShell(forcedShell)
	if err != nil {
		return nil, err
	}

	cmdLine := "minishift oc-env"
	if constants.ProfileName != profileActions.GetActiveProfile() {
		cmdLine = fmt.Sprintf("minishift oc-env --profile=%s", constants.ProfileName)
	}

	shellCfg := &OcShellConfig{
		OcDirPath: filepath.Dir(ocPath),
	}

	if noProxy {
		cmdLine = cmdLine + " --no-proxy"
		noProxyVar, noProxyValue, err := util.GetNoProxyConfig(api)
		if err != nil {
			return nil, err
		}
		shellCfg.NoProxyVar = noProxyVar
		shellCfg.NoProxyValue = noProxyValue
	}

	shellCfg.UsageHint = shell.GenerateUsageHint(userShell, cmdLine)
	shellCfg.Prefix, shellCfg.Delimiter, shellCfg.Suffix, shellCfg.PathSuffix = shell.GetPrefixSuffixDelimiterForSet(userShell)

	return shellCfg, nil
}

func executeOcTemplateStdout(shellCfg *OcShellConfig) error {
	tmpl := template.Must(template.New("envConfig").Parse(ocEnvTmpl))
	return tmpl.Execute(os.Stdout, shellCfg)
}

var ocEnvCmd = &cobra.Command{
	Use:   "oc-env",
	Short: "Sets the path of the 'oc' binary.",
	Long:  `Sets the path of OpenShift client binary 'oc'.`,
	Run: func(cmd *cobra.Command, args []string) {
		if config.InstanceStateConfig.OcPath == "" {
			atexit.ExitWithMessage(1, "Cannot find the OpenShift client binary.\nMake sure that OpenShift was provisioned successfully.")
		}

		api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
		defer api.Close()

		util.ExitIfUndefined(api, constants.MachineName)

		host, err := api.Load(constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		util.ExitIfNotRunning(host.Driver, constants.MachineName)

		var shellCfg *OcShellConfig
		ocPath := config.InstanceStateConfig.OcPath
		shellCfg, err = getOcShellConfig(api, ocPath, forceShell, noProxy)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error running the oc-env command: %s", err.Error()))
		}

		executeOcTemplateStdout(shellCfg)
	},
}

func init() {
	RootCmd.AddCommand(ocEnvCmd)
	if runtime.GOOS == "windows" {
		ocEnvCmd.Flags().BoolVar(&noProxy, "no-proxy", false, "Add the virtual machine IP to the NO_PROXY environment variable.")
	} else {
		ocEnvCmd.Flags().BoolVar(&noProxy, "no-proxy", false, "Add the virtual machine IP to the no_proxy/NO_PROXY environment variable.")
	}
	ocEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Force setting the environment for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
}
