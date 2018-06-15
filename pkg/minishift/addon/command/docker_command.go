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

type DockerCommand struct {
	*defaultCommand
}

func NewDockerCommand(command string, ignoreError bool, outputVariable string) *DockerCommand {
	defaultCommand := &defaultCommand{rawCommand: command, ignoreError: ignoreError, outputVariable: outputVariable}
	dockerCommand := &DockerCommand{defaultCommand}
	defaultCommand.fn = dockerCommand.doExecute
	return dockerCommand
}

func (c *DockerCommand) doExecute(ec *ExecutionContext, ignoreError bool, outputVariable string) error {
	commander := ec.GetDockerCommander()
	cmd := ec.Interpolate(c.rawCommand)
	fmt.Print(".")

	output, err := commander.LocalExec(cmd)
	if err != nil {
		return errors.New(fmt.Sprintf("Error executing command '%s':", err.Error()))
	}

	if outputVariable != "" {
		ec.AddToContext(outputVariable, strings.TrimSpace(output))
	}

	return nil
}
