// +build integration

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

package integration

import (
	"errors"
	"fmt"
	"github.com/minishift/minishift/test/integration/util"
	"strings"
	"sync"
	"time"
)

var commandOutputs []CommandOutput
var commandVariables []CommandVariable

type CommandOutput struct {
	Command  string
	StdOut   string
	StdErr   string
	ExitCode int
}

type CommandVariable struct {
	Name  string
	Value string
}

type Minishift struct {
	mutex  sync.Mutex
	runner util.MinishiftRunner
}

func (m *Minishift) shouldHaveState(expected string) error {
	actual := m.runner.GetStatus()
	if !strings.Contains(actual, expected) {
		return fmt.Errorf("Minishift state did not match. Expected: %s, Actual: %s", expected, actual)
	}

	return nil
}

func (m *Minishift) executingRetryingTimesWithWaitPeriodOfSeconds(command string, retry, sleep int) error {
	for i := 0; i < retry; i++ {
		err := m.executingOcCommand(command)
		if err != nil {
			return err
		}
		lastCommandOutput := getLastCommandOutput()
		if lastCommandOutput.ExitCode == 0 {
			break
		}
		time.Sleep(time.Duration(sleep) * time.Second)
	}

	return nil
}

func (m *Minishift) GetVariableByName(name string) *CommandVariable {
	if len(commandVariables) == 0 {
		return nil
	}

	for i := range commandVariables {

		variable := commandVariables[i]

		if variable.Name == name {
			return &variable
		}
	}

	return nil
}

func (m *Minishift) setVariableExecutingOcCommand(name string, command string) error {
	return m.setVariableFromExecution(name, minishift.executingOcCommand, command)
}

func (m *Minishift) SetVariable(name string, value string) {
	commandVariables = append(commandVariables,
		CommandVariable{
			name,
			value,
		})
}

func (m *Minishift) setVariableFromExecution(name string, execute commandRunner, command string) error {
	err := execute(command)
	if err != nil {
		return err
	}

	lastCommandOutput := getLastCommandOutput()
	commandFailed := (lastCommandOutput.ExitCode != 0 ||
		len(lastCommandOutput.StdErr) != 0)

	if commandFailed {
		return fmt.Errorf("Command '%s' did not execute successfully. cmdExit: %d, cmdErr: %s",
			lastCommandOutput.Command,
			lastCommandOutput.ExitCode,
			lastCommandOutput.StdErr)
	}

	m.SetVariable(name, strings.TrimSpace(lastCommandOutput.StdOut))

	return nil
}

func (m *Minishift) processVariables(command string) string {
	for _, v := range commandVariables {
		command = strings.Replace(command, fmt.Sprintf("$(%s)", v.Name), v.Value, -1)
	}
	return command
}

func (m *Minishift) executingOcCommand(command string) error {
	ocRunner := m.runner.GetOcRunner()
	if ocRunner == nil {
		return errors.New("Minishift is not Running")
	}

	command = m.processVariables(command)
	cmdOut, cmdErr, cmdExit := ocRunner.RunCommand(command)
	commandOutputs = append(commandOutputs,
		CommandOutput{
			command,
			cmdOut,
			cmdErr,
			cmdExit,
		})

	return nil
}

func (m *Minishift) executingMinishiftCommand(command string) error {
	command = m.processVariables(command)
	cmdOut, cmdErr, cmdExit := m.runner.RunCommand(command)
	commandOutputs = append(commandOutputs,
		CommandOutput{
			command,
			cmdOut,
			cmdErr,
			cmdExit,
		})

	return nil
}

func (m *Minishift) getOpenShiftUrl() string {
	cmdOut, _, _ := m.runner.RunCommand("console --url")
	return strings.TrimRight(cmdOut, "\n")
}

func (m *Minishift) getRoute(serviceName, nameSpace string) string {
	cmdOut, _, _ := m.runner.RunCommand("openshift service " + serviceName + " -n" + nameSpace + " --url")
	return strings.TrimRight(cmdOut, "\n")
}

func (m *Minishift) checkServiceRolloutForSuccess(service string, done chan bool) {
	command := fmt.Sprintf("rollout status deploymentconfig %s --watch", service)

	ocRunner := m.runner.GetOcRunner()
	cmdOut, cmdErr, cmdExit := ocRunner.RunCommand(command)
	m.mutex.Lock()
	commandOutputs = append(commandOutputs,
		CommandOutput{
			command,
			cmdOut,
			cmdErr,
			cmdExit,
		})
	m.mutex.Unlock()

	expected := "successfully rolled out"
	// if - else construct needed, else false is returned on the second time called
	if strings.Contains(cmdOut, expected) {
		done <- true
	} else {
		done <- false
	}
}

func (m *Minishift) rolloutServicesSuccessfully(servicesToCheck string) error {
	success := true
	servicesStr := strings.Replace(servicesToCheck, ", ", " ", -1)
	servicesStr = strings.Replace(servicesStr, ",", " ", -1)
	services := strings.Split(servicesStr, " ")
	total := len(services)
	done := make(chan bool, total)

	for i := 0; i < total; i++ {
		go m.checkServiceRolloutForSuccess(services[i], done)
	}

	for i := 0; i < total; i++ {
		success = success && <-done
	}

	if !success {
		return fmt.Errorf("Not all successfully rolled out")
	}
	return nil
}
