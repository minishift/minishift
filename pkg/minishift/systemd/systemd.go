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

package systemd

import (
	"fmt"

	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/minishift/systemd/action"
)

type SystemdCommander interface {
	Start(name string) (bool, error)
	Stop(name string) (bool, error)
	Restart(name string) (bool, error)
	Status(name string) (string, error)
	Enable(name string) (bool, error)
	Disable(name string) (bool, error)
	DaemonReload() (bool, error)

	// Exec runs the specified command on the Docker host
	Exec(cmd string) (string, error)
}

// SSHSystemdCommander is a SystemdCommander which communicates over ssh with the systemd daemon
type SSHSystemdCommander interface {
	SystemdCommander
	provision.SSHCommander
}

// VmSystemdCommander allows to communicate with the systemd daemon w/i the VM
type VmSystemdCommander struct {
	commander provision.SSHCommander
}

// NewVmSystemdCommander creates a new instance of a VmSystemdCommander
func NewVmSystemdCommander(sshCommander provision.SSHCommander) *VmSystemdCommander {
	return &VmSystemdCommander{
		commander: sshCommander,
	}
}

func (c VmSystemdCommander) Enable(name string) (bool, error) {
	_, err := c.service(name, action.Enable)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c VmSystemdCommander) Disable(name string) (bool, error) {
	_, err := c.service(name, action.Disable)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c VmSystemdCommander) DaemonReload() (bool, error) {
	// Might be needed for Start or Restart
	_, err := c.commander.SSHCommand("sudo systemctl daemon-reload")
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c VmSystemdCommander) Restart(name string) (bool, error) {
	c.DaemonReload()
	_, err := c.service(name, action.Restart)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c VmSystemdCommander) Start(name string) (bool, error) {
	c.DaemonReload()
	_, err := c.service(name, action.Start)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c VmSystemdCommander) Stop(name string) (bool, error) {
	_, err := c.service(name, action.Stop)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c VmSystemdCommander) Status(name string) (string, error) {
	return c.service(name, action.Status)

}

func (c VmSystemdCommander) Exec(cmd string) (string, error) {
	return c.commander.SSHCommand(cmd)
}

func (c VmSystemdCommander) service(name string, action action.Action) (string, error) {
	command := fmt.Sprintf("sudo systemctl -f %s %s", action.String(), name)

	if out, err := c.commander.SSHCommand(command); err != nil {
		return out, err
	} else {
		return out, nil
	}
}
