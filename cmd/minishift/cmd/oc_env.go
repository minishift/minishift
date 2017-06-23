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

	"github.com/minishift/minishift/pkg/minikube/constants"

	"os"
	"path/filepath"
	"text/template"

	"github.com/docker/machine/libmachine"
	cmdProfile "github.com/minishift/minishift/cmd/minishift/cmd/profile"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/minishift/minishift/pkg/util/shell"
	"github.com/spf13/cobra"
)

const (
	ocEnvTmpl = `{{ .Prefix }}PATH{{ .Delimiter }}{{ .OcDirPath }}{{ .Suffix }}{{ .UsageHint }}`
)

type OcShellConfig struct {
	shell.ShellConfig
	OcDirPath string
	UsageHint string
}

func getOcShellConfig(ocPath, forcedShell string) (*OcShellConfig, error) {
	var cmdLine string

	userShell, err := shell.GetShell(forcedShell)
	if err != nil {
		return nil, err
	}

	if cmdProfile.ProfileExists() {
		cmdLine = "minishift oc-env --profile " + cmdProfile.GetProfileName()
	} else {
		cmdLine = "minishift oc-env"
	}

	shellCfg := &OcShellConfig{
		OcDirPath: filepath.Dir(ocPath),
		UsageHint: shell.GenerateUsageHint(userShell, cmdLine),
	}

	shellCfg.Prefix, shellCfg.Suffix, shellCfg.Delimiter = shell.GetPrefixSuffixDelimiterForSet(userShell, true)

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
		if config.InstanceConfig.OcPath == "" {
			atexit.ExitWithMessage(1, "Cannot find the OpenShift client binary.\nMake sure that OpenShift was provisioned successfully.")
		}

		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		util.ExitIfUndefined(api, constants.MachineName)
		_, err := api.Load(constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		var shellCfg *OcShellConfig

		shellCfg, err = getOcShellConfig(config.InstanceConfig.OcPath, forceShell)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error running the oc-env command: %s", err.Error()))
		}
		executeOcTemplateStdout(shellCfg)
	},
}

func init() {
	RootCmd.AddCommand(ocEnvCmd)
	ocEnvCmd.Flags().StringVar(&forceShell, "shell", "", "Force setting the environment for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.")
}
