/*
Copyright (C) 2016 Red Hat, Inc.

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

package docker

import (
	"fmt"
	"github.com/docker/machine/libmachine/provision"
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/util"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type DockerCommander interface {
	// Ps returns the running containers of the Docker daemon and any error which occurred.
	Ps() (string, error)

	// Status returns the Docker status (via docker inspect) of the specified container. If the container
	// does not exist an error is returned. The valid status types are: created, restarting, running, paused
	// and exited. See also https://docs.docker.com/engine/api/v1.21/
	Status(container string) (string, error)

	// Starts the specified container. Returns true in case the start was successful, false otherwise.
	// Any occurring error is also returned.
	Start(container string) (bool, error)

	// Stops the specified container. Returns true in case the restart was successful, false otherwise.
	// Any occurring error is also returned.
	Stop(container string) (bool, error)

	// Restart restarts the specified container. Returns true in case the restart was successful, false otherwise.
	// Any occurring error is also returned.
	Restart(container string) (bool, error)

	// Cp copies a file from the Docker host to the specified destination in the specified container.
	// A successful copy will return nil. An error indicates that the copy failed.
	Cp(source string, container string, target string) error

	// Exec runs 'docker exec' with the specified options, against the specified container, using the specified
	// command and arguments. The output of the command is returned as well as any occurring error.
	Exec(options string, container string, command string, args string) (string, error)

	// LocalExec runs the specified command on the Docker host
	LocalExec(cmd string) (string, error)
}

// SSHDockerCommander is a DockerCommander which communicates over ssh with a Docker daemon
type SSHDockerCommander interface {
	DockerCommander
	provision.SSHCommander
}

// VmDockerCommander allows to communicate with the Docker daemon w/i the VM
type VmDockerCommander struct {
	commander provision.SSHCommander
}

// NewVmDockerCommander creates a new instance of a VmDockerCommander
func NewVmDockerCommander(sshCommander provision.SSHCommander) *VmDockerCommander {
	return &VmDockerCommander{
		commander: sshCommander,
	}
}

func (c VmDockerCommander) Ps() (string, error) {
	cmd := "docker ps"
	c.logCommand(cmd)
	out, err := c.commander.SSHCommand(cmd)
	return out, err
}

func (c VmDockerCommander) Start(container string) (bool, error) {
	cmd := fmt.Sprintf("docker start %s", container)
	c.logCommand(cmd)
	_, err := c.commander.SSHCommand(cmd)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c VmDockerCommander) Stop(container string) (bool, error) {
	cmd := fmt.Sprintf("docker stop %s", container)
	c.logCommand(cmd)
	_, err := c.commander.SSHCommand(cmd)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c VmDockerCommander) Restart(container string) (bool, error) {
	_, err := c.Stop(container)
	if err != nil {
		return false, err
	}

	_, err = c.Start(container)
	if err != nil {
		return false, err
	}

	// Give the container some time. In can come up and be reported 'running' just to then exit
	time.Sleep(3 * time.Second)

	retry := func() (err error) {
		status, err := c.Status(container)
		if err != nil {
			return err
		}

		if status != "running" {
			return errors.New(fmt.Sprintf("Unexpected container state '%s'", status))
		}

		return nil
	}

	err = util.RetryAfter(5, retry, 1*time.Second)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c VmDockerCommander) Cp(source string, container string, target string) error {
	cmd := fmt.Sprintf("docker cp %s %s:%s", source, container, target)
	c.logCommand(cmd)
	_, err := c.commander.SSHCommand(cmd)
	return err
}

func (c VmDockerCommander) Exec(options string, container string, command string, args string) (string, error) {
	cmd := fmt.Sprintf("docker exec %s %s %s %s", options, container, command, args)
	c.logCommand(cmd)
	return c.commander.SSHCommand(cmd)
}

func (c VmDockerCommander) LocalExec(cmd string) (string, error) {
	c.logCommand(cmd)
	out, err := c.commander.SSHCommand(cmd)
	return out, err
}

func (c VmDockerCommander) Status(container string) (string, error) {
	cmd := fmt.Sprintf("docker inspect --format='{{.State.Status}}' %s", container)
	c.logCommand(cmd)
	out, err := c.commander.SSHCommand(cmd)
	out = strings.TrimSpace(out)
	return out, err
}

func (c VmDockerCommander) logCommand(cmd string) {
	glog.V(2).Info(fmt.Sprintf("Executing docker command: '%s'", cmd))
}
