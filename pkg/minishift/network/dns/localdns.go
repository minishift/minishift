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

	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

const (
	dnsmasqPort           = "53"
	domain                = "localhost.localdomain"
	resolveFilename       = "/var/lib/minishift/resolv.dnsmasq.conf"
	additionalHostsPath   = "/var/lib/minishift/dnsmasq.hosts"
	dnsmasqContainerImage = "registry.centos.org/minishift/dnsmasq"
	dnsmasqContainerName  = "dnsmasq"
)

var (
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

// IsRunning checks whether the dnsmasq container is in running state.
func IsRunning(dockerCommander docker.DockerCommander) bool {
	status, err := dockerCommander.Status(dnsmasqContainerName)
	if err != nil || status != "running" {
		return false
	}

	return true
}

func Restart(dockerCommander docker.DockerCommander) (bool, error) {
	ok, err := dockerCommander.Restart(dnsmasqContainerName)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func Start(dockerCommander docker.DockerCommander, ipAddress string, routingDomain string) (bool, error) {
	_, err := dockerCommander.Status(dnsmasqContainerName)
	if err != nil {

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
		_, execError := dockerCommander.LocalExec(execCommand)
		if execError != nil {
			return false, execError
		}

		// matters if minikube, but let's always try
		resolvedCommand := fmt.Sprintf("sudo systemctl stop systemd-resolved")
		resolveOut, _ := dockerCommander.LocalExec(resolvedCommand)
		if resolveOut == "" {
			// when stopped, no nameservers are available
			// remove the stale symlink and write the resolv back to pull image
			dockerCommander.LocalExec(fmt.Sprintf("sudo rm -f /etc/resolv.conf; sudo cp %s /etc/resolv.conf", resolveFilename))
		}

		// container does not exist yet, we need to run first
		dnsmasqContainerRunOptions := fmt.Sprintf(dnsmasqContainerRunOptions, dnsmasqContainerName)
		_, runError := dockerCommander.Run(dnsmasqContainerRunOptions, dnsmasqContainerImage)
		if runError != nil {
			return false, runError
		}

		_, resolvError := dockerCommander.LocalExec("echo nameserver 127.0.0.1 | sudo tee /etc/resolv.conf > /dev/null")
		if execError != nil {
			return false, resolvError
		}

	} else {
		// container exists and we can start
		_, startError := dockerCommander.Start(dnsmasqContainerName)
		if startError != nil {
			return false, startError
		}
	}

	// perform host specific settings
	return handleHostDNSSettingsAfterStart(ipAddress)
}

func Stop(dockerCommander docker.DockerCommander) (bool, error) {
	_, err := dockerCommander.Stop(dnsmasqContainerName)
	if err != nil {
		return false, err
	}

	execCommand := "sudo cp /var/lib/minishift/resolv.dnsmasq.conf /etc/resolv.conf"
	_, execError := dockerCommander.LocalExec(execCommand)
	if execError != nil {
		return false, execError
	}

	return handleHostDNSSettingsAfterStop()
}

func Reset(dockerCommander docker.DockerCommander) (bool, error) {
	// remove container and configuration
	dockerCommander.Stop(dnsmasqContainerName)
	dockerCommander.LocalExec("sudo rm -rf /var/lib/minishift/dnsmasq.*; sudo rm -f /var/lib/minishift/resolv.dnsmasq.conf")
	dockerCommander.LocalExec("docker rm dnsmasq -f")

	return handleHostDNSSettingsAfterReset()
}
