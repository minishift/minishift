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

	"github.com/docker/machine/libmachine/provision"

	"github.com/minishift/minishift/pkg/util/os/atexit"
)

const (
	dnsmasqPort         = "53"
	domain              = "localhost.localdomain"
	resolveFilename     = "/var/lib/minishift/resolv.dnsmasq.conf"
	additionalHostsPath = "/var/lib/minishift/dnsmasq.hosts"
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
