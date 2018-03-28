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
	"github.com/minishift/minishift/pkg/minishift/docker"
)

// isComtainerRunning checks whether the dnsmasq container is in running state.
func isContainerRunning(sshCommander provision.SSHCommander) bool {
	dockerCommander := docker.NewVmDockerCommander(sshCommander)

	status, err := dockerCommander.Status(dnsmasqContainerName)
	if err != nil || status != "running" {
		return false
	}

	return true
}

func restartContainer(sshCommander provision.SSHCommander) (bool, error) {
	dockerCommander := docker.NewVmDockerCommander(sshCommander)

	ok, err := dockerCommander.Restart(dnsmasqContainerName)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func startContainer(sshCommander provision.SSHCommander) (bool, error) {
	dockerCommander := docker.NewVmDockerCommander(sshCommander)

	_, err := dockerCommander.Status(dnsmasqContainerName)
	if err != nil {
		// container does not exist yet, we need to run first
		dnsmasqContainerRunOptions := fmt.Sprintf(dnsmasqContainerRunOptions, dnsmasqContainerName)
		_, runError := dockerCommander.Run(dnsmasqContainerRunOptions, dnsmasqContainerImage)
		if runError != nil {
			return false, runError
		}

		// TODO: network code can add nameserver
		_, resolvError := dockerCommander.LocalExec("echo nameserver 127.0.0.1 | sudo tee /etc/resolv.conf > /dev/null")
		if resolvError != nil {
			return false, resolvError
		}
	} else {
		// container exists and we can start
		_, startError := dockerCommander.Start(dnsmasqContainerName)
		if startError != nil {
			return false, startError
		}
	}

	return isContainerRunning(sshCommander), nil
}

func stopContainer(sshCommander provision.SSHCommander) (bool, error) {
	dockerCommander := docker.NewVmDockerCommander(sshCommander)

	return dockerCommander.Stop(dnsmasqContainerName)
}

func resetContainer(sshCommander provision.SSHCommander) {
	dockerCommander := docker.NewVmDockerCommander(sshCommander)

	// remove container and configuration
	dockerCommander.Stop(dnsmasqContainerName)
	dockerCommander.LocalExec("docker rm dnsmasq -f")
}
