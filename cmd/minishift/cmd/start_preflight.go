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
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/minishift/minishift/pkg/util/filehelper"

	"github.com/docker/machine/libmachine/drivers"
	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	validations "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/minishift/shell/powershell"
	minishiftUtil "github.com/minishift/minishift/pkg/minishift/util"

	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/viper"

	cmdUtils "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minishift/constants"
	stringUtils "github.com/minishift/minishift/pkg/util/strings"
)

const (
	StorageDisk = "/mnt/sda1"
)

// preflightChecksAfterStartingHost is executed before the startHost function.
func preflightChecksBeforeStartingHost() {
	driverErrorMessage := "See the 'Setting Up the Driver Plug-in' topic (https://docs.openshift.org/latest/minishift/getting-started/setting-up-driver-plugin.html) for more information"
	prerequisiteErrorMessage := "See the 'Installing Prerequisites for Minishift' topic (https://docs.openshift.org/latest/minishift/getting-started/installing.html#install-prerequisites) for more information"

	preflightCheckSucceedsOrFails(
		configCmd.SkipCheckVMDriver.Name,
		checkVMDriver,
		fmt.Sprintf("Checking if requested hypervisor '%s' is supported on this platform", viper.GetString(configCmd.VmDriver.Name)),
		false, configCmd.WarnCheckVMDriver.Name,
		driverErrorMessage)

	switch viper.GetString(configCmd.VmDriver.Name) {
	case "xhyve":
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckXHyveDriver.Name,
			checkXhyveDriver,
			"Checking if xhyve driver is installed",
			false, configCmd.WarnCheckXHyveDriver.Name,
			driverErrorMessage)
	case "kvm":
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckKVMDriver.Name,
			checkKvmDriver,
			"Checking if KVM driver is installed",
			false, configCmd.WarnCheckKVMDriver.Name,
			driverErrorMessage)
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckKVMDriver.Name,
			checkLibvirtInstalled,
			"Checking if Libvirt is installed",
			false, configCmd.WarnCheckKVMDriver.Name,
			driverErrorMessage)
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckKVMDriver.Name,
			checkLibvirtDefaultNetworkExists,
			"Checking if Libvirt default network is present",
			false, configCmd.WarnCheckKVMDriver.Name,
			driverErrorMessage)
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckKVMDriver.Name,
			checkLibvirtDefaultNetworkActive,
			"Checking if Libvirt default network is active",
			true, configCmd.WarnCheckKVMDriver.Name,
			driverErrorMessage)
	case "hyperv":
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckHyperVDriver.Name,
			checkHypervDriverInstalled,
			"Checking if Hyper-V driver is installed",
			false, configCmd.WarnCheckHyperVDriver.Name,
			driverErrorMessage)
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckHyperVDriver.Name,
			checkHypervDriverSwitch,
			"Checking if Hyper-V driver is configured to use a Virtual Switch",
			false, configCmd.WarnCheckHyperVDriver.Name,
			driverErrorMessage)
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckHyperVDriver.Name,
			checkHypervDriverUser,
			"Checking if user is a member of the Hyper-V Administrators group",
			false, configCmd.WarnCheckHyperVDriver.Name,
			driverErrorMessage)
	case "virtualbox":
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckVBoxInstalled.Name,
			checkVBoxInstalled,
			"Checking if VirtualBox is installed",
			false, configCmd.WarnCheckVBoxInstalled.Name,
			prerequisiteErrorMessage)

	}

	preflightCheckSucceedsOrFails(
		configCmd.SkipCheckIsoUrl.Name,
		checkIsoURL,
		"Checking the ISO URL",
		false, configCmd.WarnCheckIsoUrl.Name,
		"See the 'Basic Usage' topic (https://docs.openshift.org/latest/minishift/using/basic-usage.html) for more information")
}

// preflightChecksAfterStartingHost is executed after the startHost function.
func preflightChecksAfterStartingHost(driver drivers.Driver) {
	preflightCheckSucceedsOrFailsWithDriver(
		configCmd.SkipInstanceIP.Name,
		checkInstanceIP, driver,
		"Checking for IP address",
		false, configCmd.WarnInstanceIP.Name,
		"Error determining IP address")
	preflightCheckSucceedsOrFailsWithDriver(
		configCmd.SkipCheckNetworkPing.Name,
		checkIPConnectivity, driver,
		"Checking if external host is reachable from the Minishift VM",
		true, configCmd.WarnCheckNetworkPing.Name,
		"VM is unable to ping external host")
	preflightCheckSucceedsOrFailsWithDriver(
		configCmd.SkipCheckNetworkHTTP.Name,
		checkHttpConnectivity, driver,
		"Checking HTTP connectivity from the VM",
		true, configCmd.WarnCheckNetworkHTTP.Name,
		"VM cannot connect to external URL with HTTP")
	preflightCheckSucceedsOrFailsWithDriver(
		configCmd.SkipCheckStorageMount.Name,
		checkStorageMounted, driver,
		"Checking if persistent storage volume is mounted",
		false, configCmd.WarnCheckStorageMount.Name,
		"Persistent volume storage is not mounted")
	preflightCheckSucceedsOrFailsWithDriver(
		configCmd.SkipCheckStorageUsage.Name,
		checkStorageUsage, driver,
		"Checking available disk space",
		false, configCmd.WarnCheckStorageUsage.Name,
		"Insufficient disk space on the persistent storage volume")
}

// preflightCheckFunc returns true when check passed
type preflightCheckFunc func() bool

// preflightCheckFunc used driver to interact with the VM instance and returns
// true when check passed
type preflightCheckWithDriverFunc func(driver drivers.Driver) bool

// preflightCheckSucceedsOrFails executes a pre-flight test function and prints
// the returned status in a standardized way. If the test fails and returns a
// false, the application will exit with errorMessage to describe what the
// cause is. It takes configNameOverrideIfSkipped to allow skipping the test.
// While treatAsWarning and configNameOverrideIfWarning can be used to make the
// test to be treated as a warning instead.
func preflightCheckSucceedsOrFails(configNameOverrideIfSkipped string, execute preflightCheckFunc, message string, treatAsWarning bool, configNameOverrideIfWarning string, errorMessage string) {
	fmt.Printf("-- %s ... ", message)

	isConfiguredToSkip := viper.GetBool(configNameOverrideIfSkipped)
	isConfiguredToWarn := viper.GetBool(configNameOverrideIfWarning)

	if isConfiguredToSkip {
		fmt.Println("SKIP")
		return
	}

	if execute() {
		fmt.Println("OK")
		return
	}

	fmt.Println("FAIL")
	errorMessage = fmt.Sprintf("   %s", errorMessage)
	if isConfiguredToWarn || treatAsWarning {
		fmt.Println(errorMessage)
	} else {
		atexit.ExitWithMessage(1, errorMessage)
	}
}

// preflightCheckSucceedsOrFails executes a pre-flight test function which uses
// the driver to interact with the VM instance. It prints the returned status in
// a standardized way. If the test fails and returns a false, the application
// will exit with errorMessage to describe what the cause is. It takes
// configNameOverrideIfSkipped to allow skipping the test. While treatAsWarning
// and configNameOverrideIfWarning can be used to make the test to be treated as
// a warning instead.
func preflightCheckSucceedsOrFailsWithDriver(configNameOverrideIfSkipped string, execute preflightCheckWithDriverFunc, driver drivers.Driver, message string, treatAsWarning bool, configNameOverrideIfWarning string, errorMessage string) {
	fmt.Printf("-- %s ... ", message)

	isConfiguredToSkip := viper.GetBool(configNameOverrideIfSkipped)
	isConfiguredToWarn := viper.GetBool(configNameOverrideIfWarning)

	if isConfiguredToSkip {
		fmt.Println("SKIP")
		return
	}

	if execute(driver) {
		fmt.Println("OK")
		return
	}

	fmt.Println("FAIL")
	errorMessage = fmt.Sprintf("   %s", errorMessage)
	if isConfiguredToWarn || treatAsWarning {
		fmt.Println(errorMessage)
	} else {
		atexit.ExitWithMessage(1, errorMessage)
	}
}

// checkXhyveDriver returns true if xhyve driver is available on path and has
// the setuid-bit set
func checkXhyveDriver() bool {
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
	fmt.Println("\n   Driver is available at", path)

	fmt.Printf("   Checking for setuid bit ... ")
	if fi.Mode()&os.ModeSetuid == 0 {
		return false
	}

	return true
}

// checkKvmDriver returns true if KVM driver is available on path
func checkKvmDriver() bool {
	path, err := exec.LookPath("docker-machine-driver-kvm")
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
	fmt.Println(fmt.Sprintf("\n   Driver is available at %s ... ", path))

	fmt.Printf("   Checking driver binary is executable ... ")
	if fi.Mode()&0011 == 0 {
		return false
	}
	return true
}

//checkLibvirtInstalled returns true if Libvirt is installed
func checkLibvirtInstalled() bool {
	path, err := exec.LookPath("virsh")
	if err != nil {
		return false
	}
	fi, _ := os.Stat(path)
	if fi.Mode()&os.ModeSymlink != 0 {
		path, err = os.Readlink(path)
		if err != nil {
			return false
		}
	}
	return true
}

//checkLibvirtDefaultNetworkExists returns true if the "default" network is present
func checkLibvirtDefaultNetworkExists() bool {
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "net-list")
	stdOutStdError, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	stdOut := fmt.Sprintf("%s", stdOutStdError)
	outputSlice := strings.Split(stdOut, "\n")

	for _, stdOut = range outputSlice {
		stdOut = strings.TrimSpace(stdOut)
		match, err := regexp.MatchString("^default\\s", stdOut)
		if err != nil {
			return false
		}
		if match {
			return true
		}
	}
	return false
}

//checkLibvirtDefaultNetworkActive returns true if the "default" network is active
func checkLibvirtDefaultNetworkActive() bool {
	cmd := exec.Command("virsh", "--connect", "qemu:///system", "net-list")
	cmd.Env = cmdUtils.ReplaceEnv(os.Environ(), "LC_ALL", "C")
	stdOutStdError, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	stdOut := fmt.Sprintf("%s", stdOutStdError)
	outputSlice := strings.Split(stdOut, "\n")

	for _, stdOut = range outputSlice {
		stdOut = strings.TrimSpace(stdOut)
		match, err := regexp.MatchString("^default\\s", stdOut)
		if err != nil {
			return false
		}
		if match && strings.Contains(stdOut, "active") {
			return true
		}
	}
	return false
}

// checkHypervDriverSwitch returns true if Virtual Switch has been selected
func checkHypervDriverSwitch() bool {
	switchEnv := os.Getenv("HYPERV_VIRTUAL_SWITCH")
	if switchEnv == "" {
		return false
	}

	return true
}

// checkHypervDriverInstalled returns true if Hyper-V driver is installed
func checkHypervDriverInstalled() bool {
	posh := powershell.New()

	checkIfHyperVInstalled := `@(Get-Command Get-VM).ModuleName`
	stdOut, _ := posh.Execute(checkIfHyperVInstalled)
	if !strings.Contains(stdOut, "Hyper-V") {
		return false
	}
	return true
}

// checkHypervDriverUser returns true if user is member of Hyper-V admin
func checkHypervDriverUser() bool {
	posh := powershell.New()

	// Use RID to prevent issues with localized groups: https://github.com/minishift/minishift/issues/1541
	// https://support.microsoft.com/en-us/help/243330/well-known-security-identifiers-in-windows-operating-systems
	// BUILTIN\Hyper-V Administrators => S-1-5-32-578
	checkIfMemberOfHyperVAdmins := `@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole("S-1-5-32-578")`
	stdOut, _ := posh.Execute(checkIfMemberOfHyperVAdmins)
	if !strings.Contains(stdOut, "True") {
		return false
	}

	return true
}

// checkInstanceIP makes sure the instance has an IPv4 address.
// HyperV will issue IPv6 addresses on Internal virtual switch
// https://github.com/minishift/minishift/issues/418
func checkInstanceIP(driver drivers.Driver) bool {
	ip, err := driver.GetIP()
	if err == nil && net.ParseIP(ip).To4() != nil {
		return true
	}
	return false
}

// checkIPConnectivity checks if the VM has connectivity to the outside network
func checkIPConnectivity(driver drivers.Driver) bool {
	ipToPing := viper.GetString(configCmd.CheckNetworkPingHost.Name)
	if ipToPing == "" {
		ipToPing = "8.8.8.8"
	}

	fmt.Printf("\n   Pinging %s ... ", ipToPing)
	return minishiftUtil.IsIPReachable(driver, ipToPing, false)
}

// checkHttpConnectivity allows to test outside connectivity and possible proxy support
func checkHttpConnectivity(driver drivers.Driver) bool {
	urlToRetrieve := viper.GetString(configCmd.CheckNetworkHttpHost.Name)
	if urlToRetrieve == "" {
		urlToRetrieve = "http://minishift.io/index.html"
	}

	fmt.Printf("\n   Retrieving %s ... ", urlToRetrieve)
	return minishiftUtil.IsRetrievable(driver, urlToRetrieve, false)
}

// checkStorageMounted checks if the persistent storage volume, storageDisk, is
// mounted to the VM instance
func checkStorageMounted(driver drivers.Driver) bool {
	mounted, _ := isMounted(driver, StorageDisk)
	return mounted
}

// checkStorageUsage checks if the persistent storage volume has enough storage
// space available.
func checkStorageUsage(driver drivers.Driver) bool {
	_, usedPercentage := getDiskUsage(driver, StorageDisk)
	fmt.Printf("%s used ", usedPercentage)
	usage, err := strconv.Atoi(stringUtils.GetOnlyNumbers(usedPercentage))
	if err != nil {
		return false
	}

	if usage > 80 && usage < 95 {
		fmt.Printf("!!! ")
	}
	if usage < 95 {
		return true
	}
	return false
}

// isMounted checks returns usage of mountpoint known to the VM instance
func getDiskUsage(driver drivers.Driver, mountpoint string) (string, string) {
	cmd := fmt.Sprintf(
		"df -h %s | awk 'FNR > 1 {print $2,$5}'",
		mountpoint)

	out, err := drivers.RunSSHCommandFromDriver(driver, cmd)

	if err != nil {
		return "", "ERR"
	}
	diskDetails := strings.Split(strings.Trim(out, "\n"), " ")
	diskSize := diskDetails[0]
	diskUsage := diskDetails[1]
	return diskSize, diskUsage
}

// isMounted checks if mountpoint is mounted to the VM instance
func isMounted(driver drivers.Driver, mountpoint string) (bool, error) {
	cmd := fmt.Sprintf(
		"if grep -qs %s /proc/mounts; then echo '1'; else echo '0'; fi",
		mountpoint)

	out, err := drivers.RunSSHCommandFromDriver(driver, cmd)

	if err != nil {
		return false, err
	}
	if strings.Trim(out, "\n") == "0" {
		return false, nil
	}

	return true, nil
}

// checkIsoUrl checks the Iso url and returns true if the iso file exists
func checkIsoURL() bool {
	isoUrl := viper.GetString(configCmd.ISOUrl.Name)
	for _, isoAlias := range constants.ValidIsoAliases {
		if isoUrl == isoAlias {
			return true
		}
	}

	if !strings.HasSuffix(isoUrl, ".iso") {
		return false
	}

	match, _ := regexp.MatchString(`^https?://`, isoUrl)
	if match {
		return true
	}

	if runtime.GOOS == "windows" {
		match, _ := regexp.MatchString("^file://[a-zA-Z]:/.+", isoUrl)
		if !match {
			return false
		}
		if filehelper.Exists(strings.Replace(strings.TrimPrefix(isoUrl, "file://"), "/", "\\", -1)) {
			return true
		}
	} else {
		match, _ := regexp.MatchString("^file:///.+", isoUrl)
		if !match {
			return false
		}
		if filehelper.Exists(strings.TrimPrefix(isoUrl, "file://")) {
			return true
		}
	}
	return false
}

func checkVMDriver() bool {
	err := validations.IsValidDriver(configCmd.VmDriver.Name, viper.GetString(configCmd.VmDriver.Name))
	if err != nil {
		return false
	}
	return true
}
