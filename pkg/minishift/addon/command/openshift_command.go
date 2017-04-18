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
	"github.com/minishift/minishift/pkg/minishift/openshift"
	"strings"
)

type OpenShiftCommand struct {
	*defaultCommand
}

func NewOpenShiftCommand(command string) *OpenShiftCommand {
	defaultCommand := &defaultCommand{rawCommand: command}
	openShiftCommand := &OpenShiftCommand{defaultCommand}
	defaultCommand.fn = openShiftCommand.doExecute
	return openShiftCommand
}

func (c *OpenShiftCommand) doExecute(ec *ExecutionContext) error {
	// split off the actual 'oc' command. We are using our cached oc version to run oc commands
	cmd := strings.Replace(c.rawCommand, "openshift ", "", 1)
	cmd = ec.Interpolate(cmd)
	fmt.Print(".")

	commander := ec.GetDockerCommander()
	_, err := commander.Exec("-t", openshift.OPENSHIFT_CONTAINER_NAME, openshift.OPENSHIFT_EXEC, ec.Interpolate(cmd))
	if err != nil {
		return errors.New(fmt.Sprintf("Error executing command '%s':", err.Error()))
	}

	return nil
}
