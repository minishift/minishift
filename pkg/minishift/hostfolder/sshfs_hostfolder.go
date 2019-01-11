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

package hostfolder

import (
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/golang/glog"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/minishift/hostfolder/config"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/minishift/minishift/pkg/util/os"
	"github.com/minishift/minishift/pkg/util/os/process"
	goos "os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

const (
	keyFile = "/home/docker/.ssh/id_rsa"
)

var (
	SftpPort = 2022
)

type SSHFSHostFolder struct {
	config       config.HostFolderConfig
	globalConfig *minishiftConfig.GlobalConfigType
}

func NewSSHFSHostFolder(config config.HostFolderConfig, globalConfig *minishiftConfig.GlobalConfigType) HostFolder {
	return &SSHFSHostFolder{config: config, globalConfig: globalConfig}
}

func (h *SSHFSHostFolder) Config() config.HostFolderConfig {
	return h.config
}

func (h *SSHFSHostFolder) Mount(driver drivers.Driver) error {
	if err := h.ensureRSAKeyExists(driver); err != nil {
		return err
	}
	if err := h.cacheRSAKeys(driver); err != nil {
		return err
	}

	if err := h.ensureSFTPDDaemonRunning(); err != nil {
		return err
	}

	ip, err := h.hostIP(driver)
	if err != nil {
		return err
	}

	// Mount command seems to fail occasionally. Give it a couple of attempts
	mount := func() (err error) {
		cmd := fmt.Sprintf(
			"sudo sshfs docker@%s:%s %s -o IdentityFile=%s -o 'StrictHostKeyChecking=no' -o reconnect -o allow_other -o idmap=none %s -p %d",
			ip,
			h.config.Option(config.Source),
			h.config.MountPoint(),
			keyFile,
			h.config.Option(config.ExtraOptions),
			SftpPort)

		if glog.V(2) {
			fmt.Println(cmd)
		}

		if _, err := drivers.RunSSHCommandFromDriver(driver, cmd); err != nil {
			return fmt.Errorf("error occured while mounting host folder: %s", err)
		}

		return nil
	}

	err = util.Retry(3, mount)
	if err != nil {
		errMsg := fmt.Sprintf("\nNote: Make sure that your network and firewall settings on the host allows port %d to be opened\n\n", SftpPort)
		return fmt.Errorf("%s%s", errMsg, err)
	}

	return nil
}

func (h *SSHFSHostFolder) Umount(driver drivers.Driver) error {
	cmd := fmt.Sprintf("sudo umount %s", h.config.MountPoint())

	if _, err := drivers.RunSSHCommandFromDriver(driver, cmd); err != nil {
		return fmt.Errorf("error during umounting of host folder: %s", err)
	}

	return nil
}

func (h *SSHFSHostFolder) hostIP(driver drivers.Driver) (string, error) {
	cmd := fmt.Sprint("sudo netstat -tapen | grep 'sshd: docker' | head -n1 | awk '{split($5, a, \":\"); print a[1]}'")

	out, err := drivers.RunSSHCommandFromDriver(driver, cmd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

func (h *SSHFSHostFolder) ensureSFTPDDaemonRunning() error {
	running := h.isRunning()
	if running {
		if glog.V(2) {
			fmt.Println(fmt.Sprintf("sftpd running with pid %d", h.globalConfig.SftpdPID))
		}
		return nil
	}

	sftpCmd, err := createSftpCommand()
	if err != nil {
		return err
	}

	err = sftpCmd.Start()
	if err != nil {
		return err
	}

	h.globalConfig.SftpdPID = sftpCmd.Process.Pid
	h.globalConfig.Write()
	return nil
}

func (h *SSHFSHostFolder) ensureRSAKeyExists(driver drivers.Driver) error {
	cmd := fmt.Sprintf("if [ ! -f %s ]; then ssh-keygen -t rsa -N \"\" -f %s; fi", keyFile, keyFile)
	_, err := drivers.RunSSHCommandFromDriver(driver, cmd)
	if err != nil {
		return err
	}

	return nil
}

// cacheRSAKeys will cache the private key and add the public key to authorized_keys file in profile home dir
// which will be used for authentication later during mount.
func (h *SSHFSHostFolder) cacheRSAKeys(driver drivers.Driver) error {
	cmd := fmt.Sprintf("if [ -f %s ]; then cat %s; else exit 1; fi", keyFile, keyFile)
	out, err := drivers.RunSSHCommandFromDriver(driver, cmd)
	if err != nil {
		return err
	}

	if err := filehelper.CreateOrOpenFileAndWrite(minishiftConstants.ProfilePrivateKeyPath(), out); err != nil {
		return err
	}

	cmd = fmt.Sprintf("if [ -f %s ]; then cat %s.pub; else exit 1; fi", keyFile, keyFile)
	out, err = drivers.RunSSHCommandFromDriver(driver, cmd)
	if err != nil {
		return err
	}

	// add public key to authorized keys
	if err := filehelper.CreateOrOpenFileAndWrite(minishiftConstants.ProfileAuthorizedKeysPath(), out); err != nil {
		return err
	}

	return nil
}

func (h *SSHFSHostFolder) isRunning() bool {
	if h.globalConfig.SftpdPID <= 0 {
		return false
	}

	process, err := goos.FindProcess(h.globalConfig.SftpdPID)
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

func createSftpCommand() (*exec.Cmd, error) {
	cmd, err := os.CurrentExecutable()
	if err != nil {
		return nil, err
	}

	args := []string{
		"daemon",
		"sftpd"}
	exportCmd := exec.Command(cmd, args...)
	// don't inherit any file handles
	exportCmd.Stderr = nil
	exportCmd.Stdin = nil
	exportCmd.Stdout = nil
	exportCmd.SysProcAttr = process.SysProcForBackgroundProcess()
	exportCmd.Env = process.EnvForBackgroundProcess()

	return exportCmd, nil
}
