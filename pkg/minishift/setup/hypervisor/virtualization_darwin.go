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
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
)

const (
	driverBinaryDir   = "/usr/local/bin"
	driverBinaryPath  = driverBinaryDir + "/docker-machine-driver-xhyve"
	driverDownloadUrl = "https://github.com/machine-drivers/docker-machine-driver-xhyve/releases/download/v0.3.3/docker-machine-driver-xhyve"
)

func CheckHypervisorAvailable() error {
	return nil
}

func CheckAndConfigureHypervisor() error {
	if isRoot() {
		fmt.Println("Configuring Xhyve Hypervisor ...")
		err := downloadXhyveDriver(driverBinaryPath, driverDownloadUrl)
		return err
	}
	return errors.New("This command needs to be executed as administrator or with sudo.")
}

func isXhyveConfigured() bool {
	//Following check is also present in cmd/start_preflight.go
	path, err := exec.LookPath("docker-machine-driver-xhyve")

	if err != nil {
		return false
	}

	fi, _ := os.Stat(path)
	// follow symlinks
	if fi.Mode()&os.ModeSymlink != 0 {
		path, err = os.Readlink(path)
		if err != nil {
			return false
		}
	}
	fmt.Println("Driver is available at", path)

	fmt.Print("Checking for setuid bit ... ")
	if fi.Mode()&os.ModeSetuid == 0 {
		print("FAIL")
		return false
	}
	print("OK")
	return true
}

func downloadXhyveDriver(filepath string, url string) error {
	fmt.Println("Checking if docker-machine-driver-xhyve is already present and configured ... ")
	if isXhyveConfigured() {
		return nil
	}
	os.MkdirAll(driverBinaryDir, 0751)

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	fmt.Printf("Downloading docker-machine-driver-xhyve binary to %s ... ", filepath)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	fmt.Print("OK")

	fmt.Printf("\nSetting permissions and group for %s ... ", filepath)
	err = out.Chown(0, 0)
	if err != nil {
		return err
	}

	err = out.Chmod(os.ModeSetuid | 0751)
	if err != nil {
		return err
	}
	fmt.Print("OK")

	return nil
}

func isRoot() bool {
	user, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user.")
		return false
	}
	if user.Gid == "0" {
		return true
	}
	return false
}
