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

package command

import (
	"errors"
	"fmt"
	"strings"
)

type SSHCommand struct {
	*defaultCommand
}

func NewSshCommand(command string) *SSHCommand {
	defaultCommand := &defaultCommand{rawCommand: command}
	sshCommand := &SSHCommand{defaultCommand}
	defaultCommand.fn = sshCommand.doExecute
	return sshCommand
}

func (c *SSHCommand) doExecute(ec *ExecutionContext) error {
	cmd := strings.Replace(c.rawCommand, "ssh ", "", 1)
	cmd = ec.Interpolate(cmd)
	fmt.Print(".")

	commander := ec.GetSSHCommander()
	_, err := commander.SSHCommand(ec.Interpolate(cmd))
	if err != nil {
		return errors.New(fmt.Sprintf("Error executing command '%s':", err.Error()))
	}

	return nil
}
