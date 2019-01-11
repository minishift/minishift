/*
Copyright (C) 2018 Red Hat, Inc.

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

package services

import (
	"runtime"

	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/minishift/network/proxy"
	"github.com/minishift/minishift/pkg/minishift/systemtray"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	minishiftStrings "github.com/minishift/minishift/pkg/util/strings"
	"github.com/spf13/cobra"
)

const (
	serviceNotSpecifiedError = "You must specify a service to start (use 'minishift services list' to find available services)."
)

// daemonStartCmd represents the start command
var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a minishift service.",
	Long:  "Start a minishift service.",
	Run:   runServiceStart,
}

func init() {
	ServicesCmd.AddCommand(daemonStartCmd)
}

func runServiceStart(cmd *cobra.Command, args []string) {
	if len(args) <= 0 {
		atexit.ExitWithMessage(1, serviceNotSpecifiedError)
	}
	service := args[0]
	if !minishiftStrings.Contains(minishiftConstants.ValidServices, service) {
		atexit.ExitWithMessage(1, InvalidServiceNameError)
	}

	switch service {
	case minishiftConstants.SystemtrayDaemon:
		if runtime.GOOS == "linux" {
			atexit.ExitWithMessage(0, "System tray is not supported in Linux")
		}
		minishiftTray := systemtray.NewMinishiftTray(minishiftConfig.AllInstancesConfig)
		minishiftTray.EnsureRunning()
	case minishiftConstants.SftpdDaemon:
		// Add code to start sftpd
		atexit.ExitWithMessage(0, "Start functionality for SFTP daemon is not available")
	case minishiftConstants.ProxyDaemon:
		proxy.EnsureProxyDaemonRunning()
	default:
		return
	}
}
