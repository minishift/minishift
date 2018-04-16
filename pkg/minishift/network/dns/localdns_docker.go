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
	"github.com/docker/machine/libmachine/provision"

	"github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/pkg/minishift/docker"
)

const (
	dnsmasqDefaultContainerImage = "registry.centos.org/minishift/dnsmasq"
	dnsmasqContainerName         = "dnsmasq"
)

var (
	dnsmasqContainerRunOptions = `--name %s \
	--privileged \
    -v /var/lib/minishift/dnsmasq.hosts:/etc/dnsmasq.hosts:Z \
    -v /var/lib/minishift/dnsmasq.conf:/etc/dnsmasq.conf \
    -v /var/lib/minishift/resolv.dnsmasq.conf:/etc/resolv.dnsmasq.conf \
    -p '0.0.0.0:53:53/udp' \
    -d`
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

type DockerDnsService struct {
	commander *docker.VmDockerCommander
}

func newDockerDnsService(sshCommander provision.SSHCommander) *DockerDnsService {
	return &DockerDnsService{
		commander: docker.NewVmDockerCommander(sshCommander),
	}
}

// isComtainerRunning checks whether the dnsmasq container is in running state.
func (s DockerDnsService) Status() bool {
	status, err := s.commander.Status(dnsmasqContainerName)
	if err != nil || status != "running" {
		return false
	}

	return true
}

func (s DockerDnsService) Restart() (bool, error) {
	ok, err := s.commander.Restart(dnsmasqContainerName)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (s DockerDnsService) getDnsmasqContainerImage() (string, error) {
	minishiftConfig, err := config.ReadConfig()
	if err != nil {
		return "", err
	}

	dnsmasqContainerImage := minishiftConfig[config.DnsmasqContainerImage.Name]
	if dnsmasqContainerImage != nil {
		return fmt.Sprintf("%v", dnsmasqContainerImage), nil
	}

	return dnsmasqDefaultContainerImage, nil
}

func (s DockerDnsService) Start() (bool, error) {
	_, err := s.commander.Status(dnsmasqContainerName)
	if err != nil {
		dnsmasqContainerImage, imgError := s.getDnsmasqContainerImage()
		if imgError != nil {
			return false, imgError
		}

		// container does not exist yet, we need to run first
		dnsmasqContainerRunOptions := fmt.Sprintf(dnsmasqContainerRunOptions, dnsmasqContainerName)
		_, runError := s.commander.Run(dnsmasqContainerRunOptions, dnsmasqContainerImage)
		if runError != nil {
			return false, runError
		}
	} else {
		// container exists and we can start
		_, startError := s.commander.Start(dnsmasqContainerName)
		if startError != nil {
			return false, startError
		}
	}

	return s.Status(), nil
}

func (s DockerDnsService) Stop() (bool, error) {
	return s.commander.Stop(dnsmasqContainerName)
}
