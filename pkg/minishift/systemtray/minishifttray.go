// +build !systemtray

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

package systemtray

import (
	"fmt"
	goos "os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/golang/glog"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util/os"
	"github.com/minishift/minishift/pkg/util/os/process"
)

type MinishiftTray struct {
	globalConfig *minishiftConfig.GlobalConfigType
}

func NewMinishiftTray(globalConfig *minishiftConfig.GlobalConfigType) *MinishiftTray {
	return &MinishiftTray{globalConfig: globalConfig}
}

func minishiftTrayCommand() (*exec.Cmd, error) {
	cmd, err := os.CurrentExecutable()
	if err != nil {
		return nil, err
	}
	args := []string{
		"daemon",
		"systemtray",
	}
	exportCmd := exec.Command(cmd, args...)
	// don't inherit any file handles
	exportCmd.Stderr = nil
	exportCmd.Stdin = nil
	exportCmd.Stdout = nil
	exportCmd.SysProcAttr = process.SysProcForBackgroundProcess()
	exportCmd.Env = process.EnvForBackgroundProcess()
	exportCmd.Env = append(exportCmd.Env, fmt.Sprintf("VBOX_MSI_INSTALL_PATH=%s", goos.Getenv("VBOX_MSI_INSTALL_PATH")))
	return exportCmd, nil
}

func (s *MinishiftTray) EnsureRunning() error {
	running := s.isRunning()
	if running {
		if glog.V(2) {
			fmt.Println(fmt.Sprintf("systemtray running with pid %d", s.globalConfig.SystrayPID))
		}
		return nil
	}

	trayCmd, err := minishiftTrayCommand()
	if err != nil {
		return err
	}
	err = trayCmd.Start()
	if err != nil {
		return err
	}

	s.globalConfig.SystrayPID = trayCmd.Process.Pid
	s.globalConfig.Write()
	return nil
}

func (s *MinishiftTray) isRunning() bool {
	if s.globalConfig.SystrayPID <= 0 {
		return false
	}

	process, err := goos.FindProcess(s.globalConfig.SystrayPID)
	if err != nil {
		return false
	}

	// for Windows FindProcess is enough
	if runtime.GOOS == "windows" {
		return true
	}

	// for non Windows we need to send a signal to get more information
	err = process.Signal(syscall.Signal(0))
	if err == nil {
		return true
	} else {
		return false
	}
}

func (s *MinishiftTray) GetPID() int {
	if s.isRunning() {
		return s.globalConfig.SystrayPID
	}
	return 0
}
