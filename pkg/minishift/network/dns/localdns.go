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

package dns

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/minishift/network"

	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"

	"github.com/minishift/minishift/pkg/util/os/atexit"
)

type serviceCommander interface {
	Start() (bool, error)
	Stop() (bool, error)
	Restart() (bool, error)
	Status() bool
}

// checkSupportForAddressAssignment returns true when the instance can support
// dnsmasq server from the image
// Should be using Instance state. See #1796
func checkSupportForDnsmasqServer() bool {
	if minishiftConfig.InstanceStateConfig.IsRHELBased &&
		minishiftConfig.InstanceStateConfig.SupportsDnsmasqServer {
		return true
	}
	return false
}

// isContainerized allows to force containerized deployment
// Should be using Instance config. See #1796
func isContainerized() (bool, error) {
	minishiftConfig, err := configCmd.ReadConfig()
	if err != nil || minishiftConfig[configCmd.DnsmasqContainerized.Name] == nil {
		return false, err
	}

	return minishiftConfig[configCmd.DnsmasqContainerized.Name].(bool), nil
}

func getServiceCommander(driver drivers.Driver) serviceCommander {
	sshCommander := provision.GenericSSHCommander{Driver: driver}
	isContainerized, _ := isContainerized()

	if checkSupportForDnsmasqServer() && !isContainerized {
		return newSystemdDnsService(sshCommander)
	} else {
		return newDockerDnsService(sshCommander)
	}
}

func Status(driver drivers.Driver) bool {
	return getServiceCommander(driver).Status()
}

func Start(driver drivers.Driver) (bool, error) {
	sshCommander := provision.GenericSSHCommander{Driver: driver}

	ipAddress, err := driver.GetIP()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP: %s", err.Error()))
	}

	routingSuffix := configCmd.GetDefaultRoutingSuffix(ipAddress)
	handleConfiguration(sshCommander, ipAddress, routingSuffix)

	getServiceCommander(driver).Start()

	network.AddNameserversToInstance(driver, []string{"127.0.0.1"})

	// perform host specific settings
	return handleHostDNSSettingsAfterStart(ipAddress)

}

func Stop(driver drivers.Driver) (bool, error) {
	sshCommander := provision.GenericSSHCommander{Driver: driver}

	getServiceCommander(driver).Stop()

	execCommand := "sudo cp /var/lib/minishift/resolv.dnsmasq.conf /etc/resolv.conf"
	_, execError := sshCommander.SSHCommand(execCommand)
	if execError != nil {
		return false, execError
	}

	ipAddress, err := driver.GetIP()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP: %s", err.Error()))
	}

	return handleHostDNSSettingsAfterStop(ipAddress)
}
