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
	"fmt"
	"os"
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
	ServiceNotSpecifiedError = "You need to specify a service to stop"
	InvalidServiceNameError  = "You have specified an invalid service name"
)

// daemonStopCmd represents the stop command
var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a minishift service.",
	Long:  "Stop a minishift service.",
	Run:   runServiceStop,
}

func init() {
	ServicesCmd.AddCommand(daemonStopCmd)
}

func runServiceStop(cmd *cobra.Command, args []string) {
	if len(args) <= 0 {
		atexit.ExitWithMessage(1, ServiceNotSpecifiedError)
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
		if pid := minishiftTray.GetPID(); pid > 0 {
			proc, _ := os.FindProcess(pid)
			proc.Kill()
			atexit.ExitWithMessage(0, fmt.Sprintf("Killed process with PID: %d\n", pid))
		}
	case minishiftConstants.SftpdDaemon:
		pid := minishiftConfig.AllInstancesConfig.SftpdPID
		proc, err := os.FindProcess(pid)
		if err != nil {
			atexit.ExitWithMessage(1, "Unable to get Sftp daemon process using PID")
		}
		proc.Kill()
		atexit.ExitWithMessage(0, fmt.Sprintf("Killed process with PID: %d\n", pid))
	case minishiftConstants.ProxyDaemon:
		if pid := proxy.GetPID(); pid > 0 {
			proc, _ := os.FindProcess(pid)
			proc.Kill()
			atexit.ExitWithMessage(0, fmt.Sprintf("Killed process with PID: %d\n", pid))
		}
	default:
		return
	}
}
