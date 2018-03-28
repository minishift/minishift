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
	"bytes"
	"encoding/base64"
	"fmt"
	"text/template"

	// should not be here
	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/provision"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"

	"github.com/minishift/minishift/pkg/util/os/atexit"
)

const (
	dnsmasqPort           = "53"
	domain                = "localhost.localdomain"
	resolveFilename       = "/var/lib/minishift/resolv.dnsmasq.conf"
	additionalHostsPath   = "/var/lib/minishift/dnsmasq.hosts"
	dnsmasqContainerImage = "registry.centos.org/minishift/dnsmasq"
	dnsmasqContainerName  = "dnsmasq"
	dnsmasqServiceName    = "dnsmasq"
)

var (
	dnsmasqServicePrerequisites = `sudo rm -rf /etc/dnsmasq.* /etc/resolv.dnsmasq.conf; \
sudo ln -s /var/lib/minishift/dnsmasq.hosts /etc/dnsmasq.hosts; \
sudo ln -s /var/lib/minishift/dnsmasq.conf /etc/dnsmasq.conf; \
sudo ln -s /var/lib/minishift/resolv.dnsmasq.conf /etc/resolv.dnsmasq.conf`
	dnsmasqContainerRunOptions = `--name %s \
	--privileged \
    -v /var/lib/minishift/dnsmasq.hosts:/etc/dnsmasq.hosts:Z \
    -v /var/lib/minishift/dnsmasq.conf:/etc/dnsmasq.conf \
    -v /var/lib/minishift/resolv.dnsmasq.conf:/etc/resolv.dnsmasq.conf \
    -p '0.0.0.0:53:53/udp' \
    -d` // {{.ResolveFilename}}, {{.AdditionalHostsPath}} --restart always
	dnsmasqConfigurationTemplate = `user=root
port={{.Port}}
bind-interfaces
resolv-file=/etc/resolv.dnsmasq.conf
addn-hosts=/etc/dnsmasq.hosts
expand-hosts
domain={{.Domain}}
address=/.{{.RoutingDomain}}/{{.LocalIP}}
address=/.{{.LocalIP}}.local/{{.LocalIP}}
`
)

type DnsmasqConfiguration struct {
	Port                string // 2053 (see Dockerfile)
	ResolveFilename     string // /var/lib/minishift/resolv.dnsmasq.conf
	AdditionalHostsPath string // /dnsmasq.hosts
	Domain              string // localhost.localdomain
	RoutingDomain       string // {{.LocalIP}}.{{.RoutingSuffix}}
	LocalIP             string
}

func fillDnsmasqConfiguration(dnsmasqConfiguration DnsmasqConfiguration) string {
	result := &bytes.Buffer{}

	tmpl, err := template.New("networkScript").Parse(dnsmasqConfigurationTemplate)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating network script template: %s", err.Error()))
	}
	err = tmpl.Execute(result, dnsmasqConfiguration)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error executing network script template:: %s", err.Error()))
	}

	return result.String()
}

func printDnsmasqConfiguration(dnsmasqConfiguration DnsmasqConfiguration) {
	fmt.Println(fillDnsmasqConfiguration(dnsmasqConfiguration))
}

func handleConfiguration(sshCommander provision.SSHCommander, ipAddress string, routingDomain string) (bool, error) {

	dnsmasqConfiguration := DnsmasqConfiguration{
		Port:                dnsmasqPort,
		ResolveFilename:     resolveFilename,
		AdditionalHostsPath: additionalHostsPath,
		Domain:              "minishift",
		RoutingDomain:       routingDomain,
		LocalIP:             ipAddress,
	}
	dnsmasqConfigurationFile := fillDnsmasqConfiguration(dnsmasqConfiguration) // perhaps move this to the struct as a ToString()
	encodedDnsmasqConfigurationFile := base64.StdEncoding.EncodeToString([]byte(dnsmasqConfigurationFile))
	configCommand := fmt.Sprintf(
		"echo %s | base64 --decode | sudo tee /var/lib/minishift/dnsmasq.conf > /dev/null",
		encodedDnsmasqConfigurationFile)

	execCommand := fmt.Sprintf("sudo mkdir %s && %s && sudo cp /etc/resolv.conf %s",
		additionalHostsPath,
		configCommand,
		resolveFilename)
	_, execError := sshCommander.SSHCommand(execCommand)
	if execError != nil {
		return false, execError
	}

	// matters if minikube, but let's always try
	resolvedCommand := fmt.Sprintf("sudo systemctl stop systemd-resolved")
	resolveOut, _ := sshCommander.SSHCommand(resolvedCommand)
	if resolveOut == "" {
		// when stopped, no nameservers are available
		// remove the stale symlink and write the resolv back to pull image
		sshCommander.SSHCommand(fmt.Sprintf("sudo rm -f /etc/resolv.conf; sudo cp %s /etc/resolv.conf", resolveFilename))
	}

	return true, nil
}

func IsRunning(driver drivers.Driver) bool {
	sshCommander := provision.GenericSSHCommander{Driver: driver}

	if minishiftConfig.InstanceConfig.IsRHELBased {
		return isServiceRunning(sshCommander)
	} else {
		return isContainerRunning(sshCommander)
	}
}

func Start(driver drivers.Driver) (bool, error) {
	sshCommander := provision.GenericSSHCommander{Driver: driver}

	ipAddress, err := driver.GetIP()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP: %s", err.Error()))
	}

	routingSuffix := configCmd.GetDefaultRoutingSuffix(ipAddress)
	handleConfiguration(sshCommander, ipAddress, routingSuffix)

	if minishiftConfig.InstanceConfig.IsRHELBased {
		startService(sshCommander)
	} else {
		startContainer(sshCommander)
	}

	// perform host specific settings
	return handleHostDNSSettingsAfterStart(ipAddress)

}

func Stop(driver drivers.Driver) (bool, error) {
	sshCommander := provision.GenericSSHCommander{Driver: driver}

	if minishiftConfig.InstanceConfig.IsRHELBased {
		stopService(sshCommander)
	} else {
		stopContainer(sshCommander)
	}

	execCommand := "sudo cp /var/lib/minishift/resolv.dnsmasq.conf /etc/resolv.conf"
	_, execError := sshCommander.SSHCommand(execCommand)
	if execError != nil {
		return false, execError
	}

	return handleHostDNSSettingsAfterStop()
}

func Reset(driver drivers.Driver) (bool, error) {
	sshCommander := provision.GenericSSHCommander{Driver: driver}

	if !minishiftConfig.InstanceConfig.IsRHELBased {
		resetContainer(sshCommander)
	}
	sshCommander.SSHCommand("sudo rm -rf /var/lib/minishift/dnsmasq.*; sudo rm -f /var/lib/minishift/resolv.dnsmasq.conf")

	return handleHostDNSSettingsAfterReset()
}
