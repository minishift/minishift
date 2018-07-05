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

package testsuite

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/minishift/minishift/test/integration/util"
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

type rolloutMessage struct {
	passed bool
	stdOut string
	stdErr string
}

func (m *Minishift) shouldHaveState(expected string) error {
	actual := m.runner.GetStatus()
	if !strings.Contains(actual, expected) {
		return fmt.Errorf("Minishift state did not match. Expected: %s, Actual: %s", expected, actual)
	}

	return nil
}

func (m *Minishift) profileShouldHaveState(profile string, expected string) error {
	actual := m.runner.GetProfileStatus(profile)
	if !strings.Contains(actual, expected) {
		return fmt.Errorf("Profile %s of Minishift state did not match. Expected: %s, Actual: %s", profile, expected, actual)
	}

	return nil
}

func (m *Minishift) isTheActiveProfile(profileName string) error {
	profileList := m.runner.GetProfileList()
	profileNameWtquote := strings.Replace(profileName, "\"", "", -1)
	stringArr := strings.Split(profileList, "\n")
	for i := 0; i < len(stringArr); i++ {
		if strings.Contains(stringArr[i], profileNameWtquote) && !strings.Contains(stringArr[i], "(Active)") {
			return fmt.Errorf("Profile %s is not an active profile, Actual : %s", profileName, stringArr[i])
		}
	}

	return nil
}

// containerStatus take the formatted name of container and check the status as per expectation.
// Return if any error
func (m *Minishift) containerStatus(retryCount int, retryWaitPeriod int, containerName string, expected string) error {
	var containerState string
	for i := 0; i < retryCount; i++ {
		runningContainers := m.runner.GetOpenshiftContainers(containerName)
		individualContainerRows := strings.Split(runningContainers, "\n")
		for _, individualContainer := range individualContainerRows {
			if strings.Contains(individualContainer, containerName) {
				containerId := strings.Split(individualContainer, " ")[0]
				containerState = m.runner.GetContainerStatusUsingImageId(containerId)
				if !strings.Contains(containerState, expected) {
					return fmt.Errorf("Container state did not match. Expected: %s, Actual: %s", expected, containerState)
				}
				return nil
			}
		}
		time.Sleep(time.Second * time.Duration(retryWaitPeriod))
	}
	return fmt.Errorf("Container state did not match expected state within time limit of %d seconds. Expected: %s, Actual: %s", retryCount*retryWaitPeriod, expected, containerState)
}

func (m *Minishift) executingRetryingTimesWithWaitPeriodOfSeconds(command string, retry, sleep int) error {
	for i := 0; i < retry; i++ {
		err := m.ExecutingOcCommand(command)
		if err != nil {
			return err
		}
		lastCommandOutput := GetLastCommandOutput()
		if lastCommandOutput.ExitCode == 0 {
			break
		}
		fmt.Printf("Command returned non-zero exit code: '%v', stdErr: '%v', stdOut: '%v', retrying...", lastCommandOutput.ExitCode, lastCommandOutput.StdErr, lastCommandOutput.StdOut)
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
	return m.setVariableFromExecution(name, MinishiftInstance.ExecutingOcCommand, command)
}

func (m *Minishift) setVariableExecutingMinishiftCommand(name string, command string) error {
	return m.setVariableFromExecution(name, MinishiftInstance.ExecutingMinishiftCommand, command)
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

	lastCommandOutput := GetLastCommandOutput()
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

func (m *Minishift) ExecutingOcCommand(command string) error {
	ocRunner := m.runner.GetOcRunner()
	if ocRunner == nil {
		util.LogMessage("warning", "OC binary can't be detected, minishift is not Running")
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

func (m *Minishift) ExecutingMinishiftCommand(command string) error {
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

func (m *Minishift) setImageCaching(operation string) error {
	enabled := "true"
	if operation == "disabled" {
		enabled = "false"
	}

	return m.ExecutingMinishiftCommand(fmt.Sprintf("config set image-caching %s", enabled))
}

func (m *Minishift) imageExportShouldComplete(noOfImages, maximumTime int) error {
	// poll till the output of the `minishift image list` shows number of cached images
	timeout := time.NewTimer(time.Duration(maximumTime) * time.Minute)

outerPollActive:
	for {
		select {
		case <-timeout.C:
			return errors.New("Timed out in getting the number of default cached images")
		default:
			cmdOut, _, _ := m.runner.RunCommand("image list")
			cmdOut = strings.TrimRight(cmdOut, "\n")
			numOfLines := len(strings.Split(cmdOut, "\n"))
			if numOfLines == noOfImages {
				break outerPollActive
			}
			if numOfLines > noOfImages {
				return errors.New(fmt.Sprintf("Number of expected cached images is greater than %s", noOfImages))
			}

			time.Sleep(5 * time.Second)
		}
	}

	return nil
}

func (m *Minishift) imageShouldHaveCached(image string) error {
	cmdOut, _, _ := m.runner.RunCommand("image list")
	return util.CompareExpectedWithActualMatchesRegex(image, strings.TrimRight(cmdOut, "\n"))
}

func (m *Minishift) getOpenShiftUrl() string {
	cmdOut, _, _ := m.runner.RunCommand("console --url")
	return strings.TrimRight(cmdOut, "\n")
}

func (m *Minishift) getRoute(serviceName, nameSpace string) string {
	cmdOut, _, _ := m.runner.RunCommand("openshift service " + serviceName + " -n" + nameSpace + " --url")
	return strings.TrimRight(cmdOut, "\n")
}

func (m *Minishift) checkServiceRolloutForSuccess(service string, timeout int, done chan rolloutMessage) {
	command := fmt.Sprintf("rollout status deploymentconfig %s --watch", service)

	ocRunner := m.runner.GetOcRunner()
	cmdOut, cmdErr, cmdExit := ocRunner.RunCommandWithTimeout(command, timeout)
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
		done <- rolloutMessage{passed: true, stdErr: cmdErr, stdOut: cmdOut}
	} else {
		// get application's build logs if rollout fails
		command = fmt.Sprintf("logs bc/%s", service)
		m.ExecutingOcCommand(command)

		lastCmdResult := GetLastCommandOutput()

		cmdOut += fmt.Sprintf("\n Service build output logs: %s\n", lastCmdResult.StdOut)
		cmdErr += fmt.Sprintf("\n Service build error logs: %s\n", lastCmdResult.StdErr)

		done <- rolloutMessage{passed: false, stdErr: cmdErr, stdOut: cmdOut}
	}
}

func (m *Minishift) rolloutServicesSuccessfully(servicesToCheck string) error {
	return m.rolloutServicesSuccessfullyBeforeTimeout(servicesToCheck, 0)
}

func (m *Minishift) rolloutServicesSuccessfullyBeforeTimeout(servicesToCheck string, timeout int) error {
	success := true
	var stdErrs []string
	var stdOuts []string
	servicesStr := strings.Replace(servicesToCheck, ", ", " ", -1)
	servicesStr = strings.Replace(servicesStr, ",", " ", -1)
	services := strings.Split(servicesStr, " ")
	total := len(services)
	done := make(chan rolloutMessage, total)

	for i := 0; i < total; i++ {
		go m.checkServiceRolloutForSuccess(services[i], timeout, done)
	}

	for i := 0; i < total; i++ {
		m := <-done
		stdErrs = append(stdErrs, m.stdErr)
		stdOuts = append(stdOuts, m.stdOut)
		success = success && m.passed
	}

	if !success {
		errorMessage := "Not all successfully rolled out:\n"
		for i := 0; i < total; i++ {
			errorMessage += fmt.Sprintf("Service: '%v'\n-StdOut: %v-StdErr: %v\n", services[i], stdOuts[i], stdErrs[i])
		}
		return fmt.Errorf("Not all services successfully rolled out:\n%v", errorMessage)
	}

	return nil
}
