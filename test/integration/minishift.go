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
	"time"
)

var lastCommandOutput CommandOutput

type CommandOutput struct {
	Command  string
	StdOut   string
	StdErr   string
	ExitCode int
}

type Minishift struct {
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
		if lastCommandOutput.ExitCode == 0 {
			break
		}
		time.Sleep(time.Duration(sleep) * time.Second)
	}

	return nil
}

func (m *Minishift) executingOcCommand(command string) error {
	ocRunner := m.runner.GetOcRunner()
	if ocRunner == nil {
		return errors.New("Minishift is not Running")
	}
	cmdOut, cmdErr, cmdExit := ocRunner.RunCommand(command)
	lastCommandOutput = CommandOutput{
		command,
		cmdOut,
		cmdErr,
		cmdExit,
	}

	return nil
}

func (m *Minishift) executingOcCommandSucceedsOrFails(command, expectedResult string) error {
	err := m.executingOcCommand(command)
	if err != nil {
		return err
	}
	commandFailed := (lastCommandOutput.ExitCode != 0 || len(lastCommandOutput.StdErr) != 0)
	if expectedResult == "succeeds" && commandFailed == true {
		return fmt.Errorf("Command did not execute successfully. cmdExit: %d, cmdErr: %s", lastCommandOutput.ExitCode, lastCommandOutput.StdErr)
	}
	if expectedResult == "fails" && commandFailed == false {
		return fmt.Errorf("Command executed successfully, however was expected to fail. cmdExit: %d, cmdErr: %s", lastCommandOutput.ExitCode, lastCommandOutput.StdErr)
	}

	return nil
}

func (m *Minishift) executingMinishiftCommand(command string) error {
	// TODO: there must be smarter way to destruct
	cmdOut, cmdErr, cmdExit := m.runner.RunCommand(command)
	lastCommandOutput = CommandOutput{
		command,
		cmdOut,
		cmdErr,
		cmdExit,
	}
	// Beware: you are responsible to verify the lastCommandOutput!
	return nil
}

func (m *Minishift) getOpenShiftUrl() string {
	cmdOut, _, _ := m.runner.RunCommand("console --url")
	return strings.TrimRight(cmdOut, "\n")
}
