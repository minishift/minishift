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

type EchoCommand struct {
	*defaultCommand
}

func NewEchoCommand(command string) *EchoCommand {
	defaultCommand := &defaultCommand{rawCommand: command}
	echoCommand := &EchoCommand{defaultCommand}
	defaultCommand.fn = echoCommand.doExecute
	return echoCommand
}

func (c *EchoCommand) doExecute(ec *ExecutionContext) error {
	// split off the actual 'echo' command. As we need to print rest of the string
	echo := strings.Replace(c.rawCommand, "echo ", "", 1)
	echo = ec.Interpolate(echo)

	_, err := fmt.Print("\n  " + echo)
	if err != nil {
		return errors.New(fmt.Sprintf("Error executing command '%s':", err.Error()))
	}

	return nil
}
