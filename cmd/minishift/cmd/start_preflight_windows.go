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

	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/filehelper"

	"github.com/golang/glog"
	"golang.org/x/sys/windows/registry"
)

// checkVBoxInstalled checks for the locations of VBoxManage from environment variables,
// in the default install location and in windows registry respectively, returns true
// if found in any one of the location, https://github.com/docker/machine/blob/master/drivers/virtualbox/virtualbox_windows.go
func checkVBoxInstalled() bool {
	vboxCmd := "VBoxManage.exe"
	if p := os.Getenv("VBOX_INSTALL_PATH"); p != "" {
		return filehelper.Exists(filepath.Join(p, vboxCmd))
	}
	if p := os.Getenv("VBOX_MSI_INSTALL_PATH"); p != "" {
		return filehelper.Exists(filepath.Join(p, vboxCmd))
	}
	// Look in default installation path for VirtualBox version > 5
	if filehelper.Exists(filepath.Join("C:", "Program Files", "Oracle", "VirtualBox", vboxCmd)) {
		return true
	}
	// Look in windows registry
	if p, err := findVBoxInstallDirInRegistry(); err != nil {
		return filehelper.Exists(filepath.Join(p, vboxCmd))
	}
	// fallback to search in the path and execute
	return util.CommandExecutesSuccessfully("VBoxManage", "-v")
}

// check VBOX install path in windows registry
func findVBoxInstallDirInRegistry() (string, error) {
	registryKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Oracle\VirtualBox`, registry.QUERY_VALUE)
	if err != nil {
		errorMessage := fmt.Sprintf("Can't find VirtualBox registry entries, is VirtualBox really installed properly? %s", err)
		if glog.V(2) {
			fmt.Println(errorMessage)
		}
		return "", fmt.Errorf(errorMessage)
	}
	defer registryKey.Close()

	installDir, _, err := registryKey.GetStringValue("InstallDir")
	if err != nil {
		errorMessage := fmt.Sprintf("Can't find InstallDir registry key within VirtualBox registries entries, is VirtualBox really installed properly? %s", err)
		if glog.V(2) {
			fmt.Println(errorMessage)
		}
		return "", fmt.Errorf(errorMessage)
	}

	return installDir, nil
}
