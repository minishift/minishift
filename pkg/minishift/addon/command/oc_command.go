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
	"io/ioutil"
	"os"
	"strings"
)

type OcCommand struct {
	*defaultCommand
}

func NewOcCommand(command string) *OcCommand {
	defaultCommand := &defaultCommand{rawCommand: command}
	ocCommand := &OcCommand{defaultCommand}
	defaultCommand.fn = ocCommand.doExecute
	return ocCommand
}

func (c *OcCommand) doExecute(ec *ExecutionContext) error {
	// split off the actual 'oc' command. We are using our cached oc version to run oc commands
	cmd := strings.Replace(c.rawCommand, "oc ", "", 1)
	cmd = ec.Interpolate(cmd)
	fmt.Print(".")

	commander := ec.GetOcCommander()
	exitStatus := commander.Run(ec.Interpolate(cmd), ioutil.Discard, os.Stdin)
	if exitStatus != 0 {
		return errors.New(fmt.Sprintf("Error executing command %s.", c.String()))
	}

	return nil
}
