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

package network

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine/drivers"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

const (
	configureIPAddressMessage                   = "-- Set the following network settings to VM ..."
	configureRestartNeededMessage               = "Network settings get applied to the instance on restart"
	configureIPAddressFailure                   = "Not supported on this platform or hypervisor"
	configureIPAddressAlreadySetFailure         = "Static IP address has already been assigned"
	configureNetworkNotSupportedMessage         = "The Minishift VM does not support network assignment"
	configureNetworkScriptStaticAddressTemplate = `DEVICE={{.Device}}
IPADDR={{.IPAddress}}
NETMASK={{.Netmask}}
GATEWAY={{.Gateway}}
DNS1={{.DNS1}}
DNS2={{.DNS2}}
`
	configureNetworkScriptDynamicAddressTemplate = `DEVICE={{.Device}}
USEDHCP={{.UseDHCP}}
`
	configureNetworkScriptDisabledAddressTemplate = `DEVICE={{.Device}}
DISABLED={{.Disabled}}
`
)

type NetworkSettings struct {
	Device    string
	IPAddress string
	Netmask   string
	Gateway   string
	DNS1      string
	DNS2      string
	UseDHCP   bool
	Disabled  bool
}

// checkSupportForAddressAssignment returns true when the instance can support
// minishift-set-ipaddress
func checkSupportForAddressAssignment() bool {
	if minishiftConfig.InstanceStateConfig.IsRHELBased &&
		minishiftConfig.InstanceStateConfig.SupportsNetworkAssignment {
		return true
	} else {
		atexit.ExitWithMessage(1, configureNetworkNotSupportedMessage)
	}
	return false
}

// ConfigureDynamicAssignment will write configuration files which will force DHCP
// assignment, which will be used by minishift-set-ipaddress on start of the instance.
func ConfigureDynamicAssignment(driver drivers.Driver) {
	if checkSupportForAddressAssignment() {
		fmt.Println("Writing configuration for dynamic assignment of IP address")
	}

	networkSettingsEth0 := NetworkSettings{
		Device:  "eth0",
		UseDHCP: true,
	}
	WriteNetworkSettingsToInstance(driver, networkSettingsEth0)

	networkSettingsEth1 := NetworkSettings{
		Device:  "eth1",
		UseDHCP: true,
	}
	WriteNetworkSettingsToInstance(driver, networkSettingsEth1)

	fmt.Println(configureRestartNeededMessage)
}

// ConfigureStaticAssignment will collect NetworkSettings from the current running
// instance and write this the values as configuration files, which will be used
// by minishift-set-ipaddress on start of the instance.
func ConfigureStaticAssignment(driver drivers.Driver) {
	// Not supported for KVM
	if minishiftConfig.IsKVM() {
		atexit.ExitWithMessage(1, configureIPAddressFailure)
		// related to issues with the driver (ip is retrieved from the lease)
	}

	if checkSupportForAddressAssignment() {
		fmt.Println("Writing current configuration for static assignment of IP address")
	}

	// populate the network settings struct with known values
	networkSettings := GetNetworkSettingsFromInstance(driver)

	// VirtualBox and KVM rely on two interfaces
	// eth0 is used for host communication
	// eth1 is used for the external communication
	if minishiftConfig.IsVirtualBox() {
		dhcpNetworkSettings := NetworkSettings{
			Device:  "eth0",
			UseDHCP: true,
		}
		WriteNetworkSettingsToInstance(driver, networkSettings)
		WriteNetworkSettingsToInstance(driver, dhcpNetworkSettings)
	}

	// HyperV and Xhyve rely on a single interface
	// eth0 is used for hpst and external communication
	// eth1 is disabled
	if minishiftConfig.IsHyperV() || minishiftConfig.IsXhyve() {
		disabledNetworkSettings := NetworkSettings{
			Device:   "eth1",
			Disabled: true,
		}
		WriteNetworkSettingsToInstance(driver, networkSettings)
		WriteNetworkSettingsToInstance(driver, disabledNetworkSettings)
	}

	printNetworkSettings(networkSettings)

	fmt.Println(configureRestartNeededMessage)
}

// printNetworkSettings will print to stdout the values from the struct
// NetworkSettings.
func printNetworkSettings(networkSettings NetworkSettings) {
	fmt.Println(configureIPAddressMessage)
	fmt.Println("   Device:     ", networkSettings.Device)
	fmt.Println("   IP Address: ", fmt.Sprintf("%s/%s", networkSettings.IPAddress, networkSettings.Netmask))
	if networkSettings.Gateway != "" {
		fmt.Println("   Gateway:    ", networkSettings.Gateway)
	}
	if networkSettings.DNS1 != "" || networkSettings.DNS2 != "" {
		fmt.Println("   Nameservers:", fmt.Sprintf("%s %s", networkSettings.DNS1, networkSettings.DNS2))
	}
}

// fillNetworkSettings will populate the network configuration file for use
// by the minishift-set-ipaddress script inside the VM. It takes the struct
// NetworkSettings containing the values and retuns a string containing
// a multiline config file.
func fillNetworkSettingsScript(networkSettings NetworkSettings) string {
	result := &bytes.Buffer{}

	tmpl := template.New("networkScript")

	if networkSettings.Disabled {
		tmpl.Parse(configureNetworkScriptDisabledAddressTemplate)
	} else if networkSettings.UseDHCP {
		tmpl.Parse(configureNetworkScriptDynamicAddressTemplate)
	} else {
		tmpl.Parse(configureNetworkScriptStaticAddressTemplate)
	}

	err := tmpl.Execute(result, networkSettings)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error executing network script template: %s", err.Error()))
	}

	return result.String()
}

func executeCommandOrExit(driver drivers.Driver, command string, errorMessage string) string {
	result, err := drivers.RunSSHCommandFromDriver(driver, command)

	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("%s: %s", errorMessage, err.Error()))
	}
	return result
}

// retrieve the device used for the IP address from the instance
func getInstanceNetworkDevice(driver drivers.Driver, ip string) string {
	return executeCommandOrExit(driver,
		fmt.Sprintf("ip a |grep -i '%s' | awk '{print $NF}' | tr -d '\n'", ip),
		"Error getting device")
}

// retrieve IP address and netmask from the instance
func getInstanceIPAddress(driver drivers.Driver, device string) (string, string) {
	addressInfo := executeCommandOrExit(driver,
		fmt.Sprintf("ip -o -f inet addr show %s | head -n1 | awk '/scope global/ {print $4}'", device),
		"Error getting netmask")
	ipaddress := strings.Split(strings.TrimSpace(addressInfo), "/")[0]
	netmask := strings.Split(strings.TrimSpace(addressInfo), "/")[1]

	return ipaddress, netmask
}

// get nameservers used from the instance
func getInstanceNameservers(driver drivers.Driver) []string {
	resolveInfo := executeCommandOrExit(driver,
		"cat /etc/resolv.conf |grep -i '^nameserver' | cut -d ' ' -f2 | tr '\n' ' '",
		"Error getting nameserver")
	return strings.Split(strings.TrimSpace(resolveInfo), " ")
}

// get gateway used from the instance
func getInstanceGateway(driver drivers.Driver) string {
	return executeCommandOrExit(driver,
		"route -n | grep 'UG[ \t]' | awk '{print $2}' | tr -d '\n'",
		"Error getting gateway")
}

// GetNetworkSettingsFromInstance will collect various network settings from the
// running instance and will populate a NetworkSettings struct with these values
func GetNetworkSettingsFromInstance(driver drivers.Driver) NetworkSettings {
	instanceip, err := driver.GetIP()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP address: %s", err.Error()))
	}
	if instanceip == "" {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP address: %s", "No address available"))
	}

	device := getInstanceNetworkDevice(driver, instanceip)
	ipaddress, netmask := getInstanceIPAddress(driver, device)

	if instanceip != ipaddress {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error with IP address: %s", "device has different address assigned"))
	}

	nameservers := getInstanceNameservers(driver)
	gateway := getInstanceGateway(driver)

	networkSettings := NetworkSettings{
		Device:    device,
		IPAddress: ipaddress, // ~= instanceip
		Netmask:   netmask,
		Gateway:   gateway,
	}
	if len(nameservers) > 0 {
		networkSettings.DNS1 = nameservers[0]
	}
	if len(nameservers) > 1 {
		networkSettings.DNS2 = nameservers[1]
	}

	return networkSettings
}

// WriteNetworkSettingsToInstance takes NetworkSettings and writes the values
// as a configuration file to be used by minishift-set-ipaddress
func WriteNetworkSettingsToInstance(driver drivers.Driver, networkSettings NetworkSettings) bool {
	networkScript := fillNetworkSettingsScript(networkSettings) // perhaps move this to the struct as a ToString()
	encodedScript := base64.StdEncoding.EncodeToString([]byte(networkScript))

	cmd := fmt.Sprintf(
		"echo %s | base64 --decode | sudo tee /var/lib/minishift/networking-%s > /dev/null",
		encodedScript,
		networkSettings.Device)

	if _, err := drivers.RunSSHCommandFromDriver(driver, cmd); err != nil {
		return false
	}

	return true
}

// CheckInternetConnectivity return false if user is not connected to internet
func CheckInternetConnectivity(address string) bool {
	_, err := http.Get(address)
	if err != nil {
		return false
	}
	return true
}
