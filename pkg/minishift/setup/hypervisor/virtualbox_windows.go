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

package hypervisor

import (
	"errors"
	"fmt"
	"github.com/minishift/minishift/pkg/minishift/setup/util"
	"strings"
)

var (
	vbExecutables = []string{"VirtualBox.exe", "VBoxManage.exe"}
	vbDownloadURL = "https://download.virtualbox.org/virtualbox/5.2.20/VirtualBox-5.2.20-125813-Win.exe"
)

func detectExistingInstall() (string, error) {
	out, _, err := posh.Execute(`$env:VBOX_INSTALL_PATH`)
	if err != nil {
		return "", err
	}
	if out == "" {
		out, _, err = posh.Execute(`$env:VBOX_MSI_INSTALL_PATH`)
		if err != nil {
			return "", err
		}
	}

	installLoc := strings.TrimSpace(out)
	if !util.FolderContains(installLoc, vbExecutables) {
		return "", errors.New(fmt.Sprintf("Couldn't find the VirtualBox binaries inside detected folder"))
	}

	return installLoc, nil
}

func installVirtualBox() error {
	fmt.Println("Installing VirtualBox ...")

	return nil
}

func configureVirtualBox() error {
	path, err := detectExistingInstall()
	if err != nil {
		fmt.Println("No previous installation of VirtualBox found.")
		installVirtualBox()
	}

	fmt.Printf("Previous installation of VirtualBox found at %s.", path)
	return nil
}
