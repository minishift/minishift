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
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/pkg/minikube/constants"
	validations "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/minishift/network"
	"github.com/minishift/minishift/pkg/minishift/shell/powershell"
	"github.com/minishift/minishift/pkg/util/github"

	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/viper"

	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	minishiftNetwork "github.com/minishift/minishift/pkg/minishift/network"
	openshiftVersion "github.com/minishift/minishift/pkg/minishift/openshift/version"
	stringUtils "github.com/minishift/minishift/pkg/util/strings"
)

const (
	StorageDisk                  = "/mnt/?da1"
	StorageDiskForGeneric        = "/"
	GithubAddress                = "https://github.com"
	hypervDefaultVirtualSwitchId = "c08cb7b8-9b3c-408e-8e30-5e16a3aeb444"
)

// preflightChecksBeforeStartingHost is executed before the startHost function.
func preflightChecksBeforeStartingHost() {
	if shouldPreflightChecksBeSkipped() {
		return
	}
	driverErrorMessage := "See the 'Setting Up the Virtualization Environment' topic (https://docs.openshift.org/latest/minishift/getting-started/setting-up-virtualization-environment.html) for more information"
	prerequisiteErrorMessage := "See the 'Installing Prerequisites for Minishift' topic (https://docs.openshift.org/latest/minishift/getting-started/installing.html#install-prerequisites) for more information"

	preflightCheckSucceedsOrFails(
		configCmd.SkipDeprecationCheck.Name,
		checkDeprecation,
		"Check if deprecated options are used",
		configCmd.WarnDeprecationCheck.Name,
		"")

	// Conectivity logic
	fmt.Printf("-- Checking if %s is reachable ... ", GithubAddress)
	if network.CheckInternetConnectivity(GithubAddress) {
		fmt.Printf("OK\n")
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckOpenShiftRelease.Name,
			checkOriginRelease,
			fmt.Sprintf("Checking if requested OpenShift version '%s' is valid", viper.GetString(configCmd.OpenshiftVersion.Name)),
			configCmd.WarnCheckOpenShiftRelease.Name,
			"",
		)
	} else {
		fmt.Printf("FAIL\n")
		fmt.Printf("-- Checking if requested OpenShift version '%s' is valid ... SKIP\n", viper.GetString(configCmd.OpenshiftVersion.Name))
	}
	// end of connectivity logic

	preflightCheckSucceedsOrFails(
		configCmd.SkipCheckOpenShiftVersion.Name,
		validateOpenshiftVersion,
		fmt.Sprintf("Checking if requested OpenShift version '%s' is supported", viper.GetString(configCmd.OpenshiftVersion.Name)),
		configCmd.WarnCheckOpenShiftVersion.Name,
		fmt.Sprintf("Minishift does not support OpenShift version %s. "+
			"You need to use a version >= %s\n", viper.GetString(configCmd.OpenshiftVersion.Name),
			constants.MinimumSupportedOpenShiftVersion),
	)

	preflightCheckSucceedsOrFails(
		configCmd.SkipCheckVMDriver.Name,
		checkVMDriver,
		fmt.Sprintf("Checking if requested hypervisor '%s' is supported on this platform", viper.GetString(configCmd.VmDriver.Name)),
		configCmd.WarnCheckVMDriver.Name,
		driverErrorMessage)

	switch viper.GetString(configCmd.VmDriver.Name) {
	case "xhyve":
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckXHyveDriver.Name,
			checkXhyveDriver,
			"Checking if xhyve driver is installed",
			configCmd.WarnCheckXHyveDriver.Name,
			driverErrorMessage)
	case "kvm":
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckKVMDriver.Name,
			checkKvmDriver,
			"Checking if KVM driver is installed",
			configCmd.WarnCheckKVMDriver.Name,
			driverErrorMessage)
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckKVMDriver.Name,
			checkLibvirtInstalled,
			"Checking if Libvirt is installed",
			configCmd.WarnCheckKVMDriver.Name,
			driverErrorMessage)
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckKVMDriver.Name,
			checkLibvirtDefaultNetworkExists,
			"Checking if Libvirt default network is present",
			configCmd.WarnCheckKVMDriver.Name,
			driverErrorMessage)
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckKVMDriver.Name,
			checkLibvirtDefaultNetworkActive,
			"Checking if Libvirt default network is active",
			configCmd.WarnCheckKVMDriver.Name,
			driverErrorMessage)
	case "hyperv":
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckHyperVDriver.Name,
			checkHypervDriverInstalled,
			"Checking if Hyper-V driver is installed",
			configCmd.WarnCheckHyperVDriver.Name,
			driverErrorMessage)
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckHyperVDriver.Name,
			checkHypervDriverSwitch,
			"Checking if Hyper-V driver is configured to use a Virtual Switch",
			configCmd.WarnCheckHyperVDriver.Name,
			driverErrorMessage)
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckHyperVDriver.Name,
			checkHypervDriverUser,
			"Checking if user is a member of the Hyper-V Administrators group",
			configCmd.WarnCheckHyperVDriver.Name,
			driverErrorMessage)
	case "virtualbox":
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckVBoxInstalled.Name,
			checkVBoxInstalled,
			"Checking if VirtualBox is installed",
			configCmd.WarnCheckVBoxInstalled.Name,
			prerequisiteErrorMessage)

	}

	preflightCheckSucceedsOrFails(
		configCmd.SkipCheckIsoUrl.Name,
		checkIsoURL,
		"Checking the ISO URL",
		configCmd.WarnCheckIsoUrl.Name,
		"See the 'Basic Usage' topic (https://docs.openshift.org/latest/minishift/using/basic-usage.html) for more information")
}

// preflightChecksAfterStartingHost is executed after the startHost function.
func preflightChecksAfterStartingHost(driver drivers.Driver) {
	if shouldPreflightChecksBeSkipped() {
		return
	}
	preflightCheckSucceedsOrFailsWithDriver(
		configCmd.SkipInstanceIP.Name,
		checkInstanceIP, driver,
		"Checking for IP address",
		configCmd.WarnInstanceIP.Name,
		"Error determining IP address")
	preflightCheckSucceedsOrFailsWithDriver(
		configCmd.SkipCheckNameservers.Name,
		checkNameservers, driver,
		"Checking for nameservers",
		configCmd.WarnCheckNameservers.Name,
		"VM does not have any nameserver setup")
	preflightCheckSucceedsOrFailsWithDriver(
		configCmd.SkipCheckNetworkPing.Name,
		checkIPConnectivity, driver,
		"Checking if external host is reachable from the Minishift VM",
		configCmd.WarnCheckNetworkPing.Name,
		"VM is unable to ping external host")
	preflightCheckSucceedsOrFailsWithDriver(
		configCmd.SkipCheckNetworkHTTP.Name,
		checkHttpConnectivity, driver,
		"Checking HTTP connectivity from the VM",
		configCmd.WarnCheckNetworkHTTP.Name,
		"VM cannot connect to external URL with HTTP")
	preflightCheckSucceedsOrFailsWithDriver(
		configCmd.SkipCheckStorageMount.Name,
		checkStorageMounted, driver,
		"Checking if persistent storage volume is mounted",
		configCmd.WarnCheckStorageMount.Name,
		"Persistent volume storage is not mounted")
	preflightCheckSucceedsOrFailsWithDriver(
		configCmd.SkipCheckStorageUsage.Name,
		checkStorageUsage, driver,
		"Checking available disk space",
		configCmd.WarnCheckStorageUsage.Name,
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
func preflightCheckSucceedsOrFails(configNameOverrideIfSkipped string, execute preflightCheckFunc, message string, configNameOverrideIfWarning string, errorMessage string) {
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
	if isConfiguredToWarn {
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
func preflightCheckSucceedsOrFailsWithDriver(configNameOverrideIfSkipped string, execute preflightCheckWithDriverFunc, driver drivers.Driver, message string, configNameOverrideIfWarning string, errorMessage string) {
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
	if isConfiguredToWarn {
		fmt.Println(errorMessage)
	} else {
		atexit.ExitWithMessage(1, errorMessage)
	}
}

func checkDeprecation() bool {
	// Check for deprecated options
	switchValue := os.Getenv("HYPERV_VIRTUAL_SWITCH")
	if switchValue != "" {
		fmt.Println("\n   Use of HYPERV_VIRTUAL_SWITCH has been deprecated\n   Please use: minishift config set hyperv-virtual-switch", switchValue)
		return false
	}
	return true
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
	cmd.Env = cmdUtil.ReplaceEnv(os.Environ(), "LC_ALL", "C")
	cmd.Env = cmdUtil.ReplaceEnv(cmd.Env, "LANG", "C")
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
	posh := powershell.New()

	switchName := viper.GetString(configCmd.HypervVirtualSwitch.Name)

	// check for default switch
	if switchName == "" {
		checkIfDefaultSwitchExists := fmt.Sprintf("Get-VMSwitch -Id %s | ForEach-Object { $_.Name }", hypervDefaultVirtualSwitchId)
		stdOut, stdErr, _ := posh.Execute(checkIfDefaultSwitchExists)

		if !strings.Contains(stdErr, "Get-VMSwitch") {
			// force setting the config variable
			switchName = stringUtils.ParseLines(stdOut)[0]
			viper.Set(configCmd.HypervVirtualSwitch.Name, switchName)
		}
	}

	fmt.Printf("\n   '%s' ... ", switchName)
	err := validations.IsValidHypervVirtualSwitch("hyperv-virtual-switch", switchName)
	return err == nil
}

// checkHypervDriverInstalled returns true if Hyper-V driver is installed
func checkHypervDriverInstalled() bool {
	// Check if Hyper-V's Virtual Machine Management Service is installed
	_, err := exec.LookPath("vmms.exe")
	if err != nil {
		return false
	}

	// check to see if a hypervisor is present. if hyper-v is installed and enabled,
	posh := powershell.New()

	checkHypervisorPresent := `@(Get-Wmiobject Win32_ComputerSystem).HypervisorPresent`

	stdOut, _, _ := posh.Execute(checkHypervisorPresent)
	if !strings.Contains(stdOut, "True") {
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

	//Hyper-V Administrators group check fails: https://github.com/minishift/minishift/issues/2047
	//Using SecurityIdentifier overload of IsInRole()
	checkIfMemberOfHyperVAdmins :=
		`$sid = New-Object System.Security.Principal.SecurityIdentifier("S-1-5-32-578")
	@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole($sid)`
	stdOut, _, _ := posh.Execute(checkIfMemberOfHyperVAdmins)
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

// checkNameservers will return true if the instance has nameservers
func checkNameservers(driver drivers.Driver) bool {
	return minishiftNetwork.HasNameserversConfigured(driver)
}

// checkIPConnectivity checks if the VM has connectivity to the outside network
func checkIPConnectivity(driver drivers.Driver) bool {
	ipToPing := viper.GetString(configCmd.CheckNetworkPingHost.Name)

	fmt.Printf("\n   Pinging %s ... ", ipToPing)
	return minishiftNetwork.IsIPReachable(driver, ipToPing, false)
}

// checkHttpConnectivity allows to test outside connectivity and possible proxy support
func checkHttpConnectivity(driver drivers.Driver) bool {
	urlToRetrieve := viper.GetString(configCmd.CheckNetworkHttpHost.Name)

	fmt.Printf("\n   Retrieving %s ... ", urlToRetrieve)
	return minishiftNetwork.IsRetrievable(driver, urlToRetrieve, false)
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
	_, usedPercentage, _ := getDiskUsage(driver, StorageDisk)
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
func getDiskUsage(driver drivers.Driver, mountpoint string) (string, string, string) {
	cmd := fmt.Sprintf(
		"df -h %s | awk 'FNR > 1 {print $2,$5,$6}'",
		mountpoint)

	out, err := drivers.RunSSHCommandFromDriver(driver, cmd)

	if err != nil {
		return "", "ERR", ""
	}
	diskDetails := strings.Split(strings.Trim(out, "\n"), " ")
	diskSize := diskDetails[0]
	diskUsage := diskDetails[1]
	diskMountPoint := diskDetails[2]
	return diskSize, diskUsage, diskMountPoint
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
	err := validations.IsValidISOUrl(configCmd.ISOUrl.Name, isoUrl)
	if err != nil {
		return false
	}
	return true
}

func checkVMDriver() bool {
	err := validations.IsValidDriver(configCmd.VmDriver.Name, viper.GetString(configCmd.VmDriver.Name))
	if err != nil {
		return false
	}
	return true
}

func validateOpenshiftVersion() bool {
	requestedVersion := viper.GetString(configCmd.OpenshiftVersion.Name)

	valid, err := openshiftVersion.IsGreaterOrEqualToBaseVersion(requestedVersion, constants.MinimumSupportedOpenShiftVersion)
	if err != nil {
		return false
	}

	if !valid {
		return false
	}
	return true
}

// checkOriginRelease return true if specified version of OpenShift is released
func checkOriginRelease() bool {
	client := github.Client()
	_, _, err := client.Repositories.GetReleaseByTag("openshift", "origin", viper.GetString(configCmd.OpenshiftVersion.Name))
	if err != nil && github.IsRateLimitError(err) {
		fmt.Println("\n   Hit github rate limit:", err)
		return false
	}

	if err != nil {
		fmt.Printf("%s is not a valid OpenShift version", viper.GetString(configCmd.OpenshiftVersion.Name))
		return false
	}
	return true
}
