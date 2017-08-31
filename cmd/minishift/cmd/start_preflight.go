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
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	miniutil "github.com/minishift/minishift/pkg/minishift/util"

	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/viper"
)

const (
	StorageDisk = "/mnt/sda1"
)

// preflightChecksAfterStartingHost is executed before the startHost function.
func preflightChecksBeforeStartingHost() {
	switch viper.GetString(configCmd.VmDriver.Name) {
	case "xhyve":
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckXHyveDriver.Name,
			checkXhyveDriver,
			"Checking if xhyve driver is installed",
			false, configCmd.WarnCheckXHyveDriver.Name,
			"See the 'Setting Up the Driver Plug-in' topic for more information")
	case "kvm":
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckKVMDriver.Name,
			checkKvmDriver,
			"Checking if KVM driver is installed",
			false, configCmd.WarnCheckXHyveDriver.Name,
			"See the 'Setting Up the Driver Plug-in' topic for more information")
	case "hyperv":
		preflightCheckSucceedsOrFails(
			configCmd.SkipCheckHyperVDriver.Name,
			checkHypervDriver,
			"Checking if Hyper-V driver is configured",
			false, configCmd.WarnCheckHyperVDriver.Name,
			"Hyper-V virtual switch is not set")
	}
}

// preflightChecksAfterStartingHost is executed after the startHost function.
func preflightChecksAfterStartingHost(driver drivers.Driver) {
	preflightCheckSucceedsOrFailsWithDriver(
		configCmd.SkipInstanceIP.Name,
		checkInstanceIP, driver,
		"Checking for IP address",
		false, configCmd.WarnInstanceIP.Name,
		"Error determining IP address")
	/*
		// This happens too late in the preflight, as provisioning needs an IP already
			preflightCheckSucceedsOrFailsWithDriver(
				configCmd.SkipCheckNetworkHost.Name,
				checkVMConnectivity, driver,
				"Checking if VM is reachable from host",
				configCmd.WarnCheckNetworkHost.Name,
				"Please check our troubleshooting guide")
	*/
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

// checkHypervDriver returns true if Virtual Switch has been selected
func checkHypervDriver() bool {
	switchEnv := os.Getenv("HYPERV_VIRTUAL_SWITCH")
	if switchEnv == "" {
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

// checkVMConnectivity checks if VM instance IP is reachable from the host
func checkVMConnectivity(driver drivers.Driver) bool {
	// used to check if the host can reach the VM
	ip, _ := driver.GetIP()

	cmd := exec.Command("ping", "-n 1", ip)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	fmt.Printf("%s\n", stdoutStderr)
	return false
}

// checkIPConnectivity checks if the VM has connectivity to the outside network
func checkIPConnectivity(driver drivers.Driver) bool {
	ipToPing := viper.GetString(configCmd.CheckNetworkPingHost.Name)
	if ipToPing == "" {
		ipToPing = "8.8.8.8"
	}

	fmt.Printf("\n   Pinging %s ... ", ipToPing)
	return miniutil.IsIPReachable(driver, ipToPing, false)
}

// checkHttpConnectivity allows to test outside connectivity and possible proxy support
func checkHttpConnectivity(driver drivers.Driver) bool {
	urlToRetrieve := viper.GetString(configCmd.CheckNetworkHttpHost.Name)
	if urlToRetrieve == "" {
		urlToRetrieve = "http://minishift.io/index.html"
	}

	fmt.Printf("\n   Retrieving %s ... ", urlToRetrieve)
	return miniutil.IsRetrievable(driver, urlToRetrieve, false)
}

// checkStorageMounted checks if the peristent storage volume, storageDisk, is
// mounted to the VM instance
func checkStorageMounted(driver drivers.Driver) bool {
	mounted, _ := isMounted(driver, StorageDisk)
	return mounted
}

// checkStorageUsage checks if the peristent storage volume has enough storage
// space available.
func checkStorageUsage(driver drivers.Driver) bool {
	usedPercentage := getDiskUsage(driver, StorageDisk)
	fmt.Printf("%s ", usedPercentage)
	usedPercentage = strings.TrimRight(usedPercentage, "%")
	usage, err := strconv.ParseInt(usedPercentage, 10, 8)
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
func getDiskUsage(driver drivers.Driver, mountpoint string) string {
	cmd := fmt.Sprintf(
		"df -h %s | awk 'FNR > 1 {print $5}'",
		mountpoint)

	out, err := drivers.RunSSHCommandFromDriver(driver, cmd)

	if err != nil {
		return "ERR"
	}

	return strings.Trim(out, "\n")
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
