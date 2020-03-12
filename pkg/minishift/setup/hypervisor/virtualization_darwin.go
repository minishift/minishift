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
	driverBinaryDir     = "/usr/local/bin"
	driverBinaryPath    = driverBinaryDir + "/docker-machine-driver-hyperkit"
	hyperkitBinaryPath  = driverBinaryDir + "/hyperkit"
	driverDownloadUrl   = "https://github.com/machine-drivers/docker-machine-driver-hyperkit/releases/download/v1.0.0/docker-machine-driver-hyperkit"
	hyperkitDownloadUrl = "https://github.com/code-ready/machine-driver-hyperkit/releases/download/v0.12.6/hyperkit"
)

func CheckHypervisorAvailable() error {
	return nil
}

func CheckAndConfigureHypervisor() error {
	if isRoot() {
		fmt.Println("Configuring Hyperkit Hypervisor ...")
		err := downloadHyperkit(hyperkitBinaryPath, hyperkitDownloadUrl)
		if err != nil {
			return err
		}
		return downloadHyperkitDriver(driverBinaryPath, driverDownloadUrl)
	}
	return errors.New("This command needs to be executed as administrator or with sudo.")
}

func isHyperkitDriverConfigured() bool {
	//Following check is also present in cmd/start_preflight.go
	path, err := exec.LookPath("docker-machine-driver-hyperkit")
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
	fmt.Println("\nDriver is available at", path)
	fmt.Print("Checking for setuid bit ... ")
	if fi.Mode()&os.ModeSetuid == 0 {
		fmt.Print("FAIL")
		return false
	}
	fmt.Print("OK")
	return true
}

func isHyperkitConfigured() bool {
	//Check if hyperkit binary is present
	path, err := exec.LookPath("hyperkit")
	if err != nil {
		return false
	}
	// follow symlink
	fi, _ := os.Stat(path)
	if fi.Mode()&os.ModeSymlink != 0 {
		path, err = os.Readlink(path)
		if err != nil {
			return false
		}
	}
	fmt.Println("\nHyperkit is available at", path)
	fmt.Print("Checking for setuid bit ... ")
	if fi.Mode()&os.ModeSetuid == 0 {
		fmt.Print("FAIL")
		return false
	}
	fmt.Print("OK")
	return true
}

func downloadHyperkitDriver(filepath string, url string) error {
	fmt.Print("\nChecking if docker-machine-driver-hyperkit is already present and configured ... ")
	if isHyperkitDriverConfigured() {
		return nil
	}
	fmt.Print("FAIL")
	fmt.Print("\nDownloading docker-machine-driver-hyperkit to: ", filepath, " ... ")
	os.MkdirAll(driverBinaryDir, 0751)
	err := download(url, filepath)
	if err != nil {
		fmt.Print("FAIL")
		return err
	}
	fmt.Print("OK")
	return nil
}

func downloadHyperkit(filepath string, url string) error {
	fmt.Print("Checking if Hyperkit is already present ... ")
	if isHyperkitConfigured() {
		return nil
	}
	fmt.Print("FAIL")
	fmt.Print("\nDownloading Hyperkit to: ", filepath, " ... ")
	os.MkdirAll(driverBinaryDir, 0751)
	err := download(url, filepath)
	if err != nil {
		fmt.Print("FAIL")
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

func download(url, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	err = out.Chown(0, 0)
	if err != nil {
		return err
	}

	return out.Chmod(os.ModeSetuid | 0751)
}
