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

package parser

import (
	"fmt"
	"strings"

	"github.com/minishift/minishift/pkg/minishift/addon/command"
	"github.com/pkg/errors"
)

const (
	dockerCommand    = "docker"
	ocCommand        = "oc"
	openShiftCommand = "openshift"
	sleepCommand     = "sleep"
	sshCommand       = "ssh"
	echoCommand      = "echo"
)

type CommandHandler interface {
	// Create attempts to parse the given string into a Command instance of its type. In case
	// s does not represent a command which can be handled by this Command instance, Create of the specified next
	// Command is called. If there is no next Command, nil is returned.
	Handle(next CommandHandler, s string) (command.Command, error)

	// Parse attempts to parse the given string into a Command instance of its type. nil is returned in case
	// s does not represent a command which can be handled by this Command instance.
	Parse(s string) command.Command
}

type defaultCommandHandler struct {
	CommandHandler

	next CommandHandler
}

func (dc *defaultCommandHandler) SetNext(next CommandHandler) {
	dc.next = next
}

func (dc *defaultCommandHandler) Handle(c CommandHandler, s string) (command.Command, error) {
	newCommand := c.Parse(s)
	if newCommand != nil {
		return newCommand, nil
	} else if dc.next != nil {
		return dc.next.Handle(dc.next, s)
	} else {
		return nil, errors.New(fmt.Sprintf("Unable to process command: '%s'", s))
	}
}

type DockerCommandHandler struct {
	*defaultCommandHandler
}

func (c *DockerCommandHandler) Parse(s string) command.Command {
	if strings.HasPrefix(s, dockerCommand) {
		return command.NewDockerCommand(s)
	}
	return nil
}

type OcCommandHandler struct {
	*defaultCommandHandler
}

func (c *OcCommandHandler) Parse(s string) command.Command {
	if strings.HasPrefix(s, ocCommand) {
		return command.NewOcCommand(s)
	}
	return nil
}

type OpenShiftCommandHandler struct {
	*defaultCommandHandler
}

func (c *OpenShiftCommandHandler) Parse(s string) command.Command {
	if strings.HasPrefix(s, openShiftCommand) {
		return command.NewOpenShiftCommand(s)
	}
	return nil
}

type SleepCommandHandler struct {
	*defaultCommandHandler
}

func (c *SleepCommandHandler) Parse(s string) command.Command {
	if strings.HasPrefix(s, sleepCommand) {
		return command.NewSleepCommand(s)
	}
	return nil
}

type SSHCommandHandler struct {
	*defaultCommandHandler
}

func (c *SSHCommandHandler) Parse(s string) command.Command {
	if strings.HasPrefix(s, sshCommand) {
		return command.NewSshCommand(s)
	}
	return nil
}

type EchoCommandHandler struct {
	*defaultCommandHandler
}

func (c *EchoCommandHandler) Parse(s string) command.Command {
	if strings.HasPrefix(s, echoCommand) {
		return command.NewEchoCommand(s)
	}
	return nil
}
