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

package hostfolder

import (
	"errors"
	"fmt"
	"github.com/docker/machine/libmachine/drivers"

	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/minishift/hostfolder/config"
	"github.com/minishift/minishift/pkg/minishift/network"
	"github.com/minishift/minishift/pkg/util"
	"strings"
)

type CifsHostFolder struct {
	config config.HostFolderConfig
}

func NewCifsHostFolder(config config.HostFolderConfig) HostFolder {
	return &CifsHostFolder{config: config}
}

func (h *CifsHostFolder) Config() config.HostFolderConfig {
	return h.config
}

func (h *CifsHostFolder) Mount(driver drivers.Driver) error {
	// If "Users" is used as name, determine the IP of host for UNC path on startup
	if h.config.Name == "Users" {
		hostIP, _ := network.DetermineHostIP(driver)
		h.config.Options[config.UncPath] = fmt.Sprintf("//%s/Users", hostIP)
	}

	print(fmt.Sprintf("   Mounting '%s': '%s' as '%s' ... ",
		h.config.Name,
		h.config.Options[config.UncPath],
		h.config.MountPoint()))

	if !h.isCifsHostReachable(driver) {
		fmt.Print("Unreachable\n")
		return errors.New("host folder is unreachable")
	}

	password, err := util.DecryptText(h.config.Options[config.Password])
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf(
		"sudo mount -t cifs %s %s -o username=%s,password=%s",
		h.config.Options[config.UncPath],
		h.config.MountPoint(),
		h.config.Options[config.UserName],
		password)

	if minishiftConfig.InstanceConfig.IsRHELBased {
		cmd = fmt.Sprintf("%s,context=system_u:object_r:svirt_sandbox_file_t:s0", cmd)
	}

	if len(h.config.Options[config.Domain]) > 0 {
		cmd = fmt.Sprintf("%s,domain=%s", cmd, h.config.Options["domain"])
	}

	if err := h.ensureMountPointExists(driver); err != nil {
		fmt.Println("FAIL")
		return fmt.Errorf("error occured while creating mountpoint. %s", err)
	}

	if _, err := drivers.RunSSHCommandFromDriver(driver, cmd); err != nil {
		fmt.Println("FAIL")
		return fmt.Errorf("error occured while mounting host folder: %s", err)
	} else {
		fmt.Println("OK")
	}

	return nil
}

func (h *CifsHostFolder) Umount(driver drivers.Driver) error {
	cmd := fmt.Sprintf(
		"sudo umount %s",
		h.config.MountPoint())

	if _, err := drivers.RunSSHCommandFromDriver(driver, cmd); err != nil {
		fmt.Println("FAIL")
		return fmt.Errorf("error during umounting of host folder: %s", err)
	} else {
		fmt.Println("OK")
	}

	return nil
}

func (h *CifsHostFolder) isCifsHostReachable(driver drivers.Driver) bool {
	uncPath := h.config.Options[config.UncPath]

	host := ""

	splitHost := strings.Split(uncPath, "/")
	if len(splitHost) > 2 {
		host = splitHost[2]
	}

	if host == "" {
		return false
	}

	return network.IsIPReachable(driver, host, false)
}

func (h *CifsHostFolder) ensureMountPointExists(driver drivers.Driver) error {
	cmd := fmt.Sprintf(
		"sudo mkdir -p %s",
		h.config.MountPoint())

	if _, err := drivers.RunSSHCommandFromDriver(driver, cmd); err != nil {
		return err
	}

	return nil
}
