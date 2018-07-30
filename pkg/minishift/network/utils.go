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

package network

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/minishift/shell/powershell"
)

var VMSwitch string

// This will return the address as used by libmachine
func GetIP(driver drivers.Driver) (string, error) {
	ip, err := driver.GetIP()
	if err != nil {
		return "", err
	}
	return ip, nil
}

func DetermineHostIP(driver drivers.Driver) (string, error) {
	// In case of generic driver we don't need to get the bridge network
	// since it's by default ping able from host IP.
	if driver.DriverName() == "generic" {
		return "", nil
	}
	instanceIP, err := driver.GetIP()
	if err != nil {
		return "", err
	}

	for _, hostaddr := range HostIPs() {

		if NetworkContains(hostaddr, instanceIP) {
			hostip, _, _ := net.ParseCIDR(hostaddr)
			// This step is not working with Windows + VirtualBox as of now
			// This test is required for CIFS mount-folder case.
			// Details: https://github.com/minishift/minishift/issues/2561
			/*if IsIPReachable(driver, hostip.String(), false) {
				return hostip.String(), nil
			}*/
			return hostip.String(), nil
		}
	}

	return "", errors.New("unknown error occurred")
}

// HasNameserversConfigured returns true if the instance uses nameservers
// This is related to an issues when LCOW is used on Windows.
func HasNameserversConfigured(driver drivers.Driver) bool {
	cmd := "cat /etc/resolv.conf | grep -i '^nameserver' | wc -l | tr -d '\n'"
	out, err := drivers.RunSSHCommandFromDriver(driver, cmd)

	if err != nil {
		return false
	}

	i, _ := strconv.Atoi(out)

	return i != 0
}

// AddHostEntryToInstance will add an entry to /etc/hosts
func AddHostEntryToInstance(driver drivers.Driver, hostname string, ipAddress string) {
	executeCommandOrExit(driver, fmt.Sprintf("echo '%s %s' | sudo tee -a /etc/hosts",
		ipAddress, hostname),
		"Error adding host entry to instance")
}

// AddNameserversToInstance will add additional nameservers to the end of the
// /etc/resolv.conf file inside the instance.
func AddNameserversToInstance(driver drivers.Driver, nameservers []string) {
	// TODO: verify values to be valid

	for _, ns := range nameservers {
		addNameserverToInstance(driver, ns)
	}
}

// writes nameserver to the /etc/resolv.conf inside the instance
func addNameserverToInstance(driver drivers.Driver, nameserver string) {
	executeCommandOrExit(driver,
		fmt.Sprintf("NS=%s; cat /etc/resolv.conf |grep -i \"^nameserver $NS\" || echo \"nameserver $NS\" | sudo tee -a /etc/resolv.conf", nameserver),
		"Error adding nameserver")
}

func HasNameserverConfiguredLocally(nameserver string) (bool, error) {
	file, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		return false, err
	}

	return strings.Contains(string(file), nameserver), nil
}

// AllowInsecureCertificatesOnLocalConnections will not verify certificates for TLS connections
func OverrideInsecureSkipVerifyForLocalConnections(insecureSkipVerify bool) error {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify}
	// We need to force a connection, else this will not work
	_, err := http.Get("http://localhost")
	return err
}

// OverrideProxyForLocalConnections will set a single proxy target for the default connection
func OverrideProxyForLocalConnections(proxyAddr string) error {
	proxy := func(*http.Request) (*url.URL, error) {
		u, _ := url.Parse(proxyAddr)
		return u, nil
	}
	http.DefaultTransport.(*http.Transport).Proxy = proxy
	// We need to force a connection, else this will not work
	_, err := http.Get("http://localhost")
	return err
}

// IsUsingDefaultSwitch returns true if the Default Switch is used before creating the VM
func IsUsingDefaultSwitch() bool {
	posh := powershell.New()

	checkIfDefaultSwitchExists := fmt.Sprintf("Get-VMSwitch -Id %s | ForEach-Object { $_.Name }", minishiftConstants.HypervDefaultVirtualSwitchId)
	stdOut, stdErr, _ := posh.Execute(checkIfDefaultSwitchExists)

	// If the config variable is epmty and default switch exists
	// If stdErr contains the command then execution failed
	if VMSwitch == "" && !strings.Contains(stdErr, "Get-VMSwitch") {
		return true
	}
	// if the config variable was set to use the default switch
	if VMSwitch == strings.TrimSpace(stdOut) {
		return true
	}
	return false
}

// GetVMSwitchId returns the ID of the vswitch in use by the running minishift vm on hyperv
func GetVMSwitchId() string {
	posh := powershell.New()

	switchIdCommand := fmt.Sprintf("(Get-VMSwitch (Get-VM \"%s\" | Get-VMNetworkAdapter | select SwitchName).SwitchName).Id.Guid", profile.GetActiveProfile())
	stdOut, _, _ := posh.Execute(switchIdCommand)

	return strings.TrimSpace(stdOut)
}

// GetVMSwitchName returns the name of the vswitch in use by the running minishift vm on hyperv
func GetVMSwitchName() string {
	posh := powershell.New()

	switchNameCommand := fmt.Sprintf("(Get-VM \"%s\" | Get-VMNetworkAdapter | select SwitchName).SwitchName", profile.GetActiveProfile())
	stdOut, _, _ := posh.Execute(switchNameCommand)

	return strings.TrimSpace(stdOut)
}
