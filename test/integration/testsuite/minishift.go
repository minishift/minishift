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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/minishift/minishift/test/integration/util"
)

var commandOutputs []CommandOutput

type CommandOutput struct {
	Command  string
	StdOut   string
	StdErr   string
	ExitCode int
}

type Minishift struct {
	mutex  sync.Mutex
	runner util.MinishiftRunner
}

type rolloutMessage struct {
	stdOut   string
	stdErr   string
	exitCode int
	err      error
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

func (m *Minishift) ShouldHaveNoOfProcessors(noOfprocessor int) error {
	cpuInfo, err := m.runner.CpuInfo()
	if err != nil {
		return err
	}
	cpuString := "processor\\s*:\\s*\\d"
	re := regexp.MustCompile(cpuString)
	listItems := re.FindAllString(cpuInfo, -1)
	if len(listItems) != noOfprocessor {
		return fmt.Errorf("The vm is running with %d no. of cpus. Expected: %d", len(listItems), noOfprocessor)
	}
	return nil
}

func (m *Minishift) ShouldHaveDiskSize(minDiskSize int, maxDiskSize int) error {
	diskInfo, err := m.runner.DiskInfo()
	if err != nil {
		return err
	}
	re := regexp.MustCompile("Disk\\s*\\/dev\\/.da:\\s*(\\d+\\.[0-9]{0,2})\\s*(GB|GiB)")
	matches := re.FindStringSubmatch(diskInfo)
	if matches == nil {
		return fmt.Errorf("Unable to find disk size string from command output: %s", diskInfo)
	}
	diskSize, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return fmt.Errorf("Unable to parse disk size %v GB to float number. Error: %v", matches[1], err)
	}
	if diskSize >= float64(minDiskSize) && diskSize <= float64(maxDiskSize) {
		return nil
	}
	return fmt.Errorf("The vm is running with disk size of %v. Expected range : %d GB - %d GB", matches[1], minDiskSize, maxDiskSize)
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
func (m *Minishift) containerStatus(retryCount int, waitTime string, containerName string, expected string) error {
	var containerState string
	var err error
	waitDuration, err := time.ParseDuration(waitTime)
	if err != nil {
		return err
	}

	for i := 0; i < retryCount; i++ {
		var err error
		containerID, err := m.runner.GetContainerID(containerName)
		if err != nil {
			fmt.Printf("error getting containers information: '%v', retrying in %v...\n", err, waitTime)
			time.Sleep(waitDuration)
			continue
		}

		containerState, err = m.runner.GetContainerStatusUsingImageID(containerID)
		if err != nil {
			fmt.Printf("error getting state of container '%v' with ID '%v' - error: '%v', retrying in %v...\n", containerName, containerID, err, waitTime)
			time.Sleep(waitDuration)
			continue
		}

		if !strings.Contains(containerState, expected) {
			fmt.Printf("container state does not match, expected: '%v', actual: '%v', retrying in %v...\n", expected, containerState, waitTime)
			time.Sleep(waitDuration)
			continue
		}

		return nil
	}

	return fmt.Errorf("Container state did not match expected state within time limit. Expected: %s, Actual: %s", expected, containerState)
}

func (m *Minishift) executingRetryingTimesWithWaitPeriodOfTime(command string, retryCount int, retryTime string) error {
	retryDuration, err := time.ParseDuration(retryTime)
	if err != nil {
		return err
	}

	for count := 0; count < retryCount; count++ {
		err := m.ExecutingOcCommand(command)
		if err != nil {
			return err
		}
		lastCommandOutput := GetLastCommandOutput()
		if lastCommandOutput.ExitCode == 0 {
			return nil
		}
		fmt.Printf("execution (%v) failed, exit code: '%v', stdErr: '%v', stdOut: '%v', retrying in %v...\n",
			count+1,
			lastCommandOutput.ExitCode,
			lastCommandOutput.StdErr,
			lastCommandOutput.StdOut,
			retryTime)
		time.Sleep(retryDuration)
	}

	return fmt.Errorf("command did not run successfully within %v retries", retryCount)
}

func (m *Minishift) setVariableExecutingOcCommand(name string, command string) error {
	return m.setVariableFromExecution(name, MinishiftInstance.ExecutingOcCommand, command)
}

func (m *Minishift) setVariableExecutingMinishiftCommand(name string, command string) error {
	return m.setVariableFromExecution(name, MinishiftInstance.ExecutingMinishiftCommand, command)
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

	util.SetVariable(name, strings.TrimSpace(lastCommandOutput.StdOut))

	return nil
}

func (m *Minishift) ExecutingOcCommand(command string) error {
	ocRunner := m.runner.GetOcRunner()
	if ocRunner == nil {
		util.LogMessage("warning", "oc binary can't be used, Minishift is not Running")
		return errors.New("oc binary can't be used, Minishift is not Running")
	}

	command = util.ProcessVariables(command)
	cmdOut, cmdErr, cmdExit, err := ocRunner.RunCommand(command)
	if err != nil {
		return err
	}

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
	command = util.ProcessVariables(command)
	cmdOut, cmdErr, cmdExit, err := m.runner.RunCommand(command)
	if err != nil {
		return err
	}

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

func (m *Minishift) imageExportShouldComplete(noOfImages int, maximumTime string) error {
	// poll till the output of the `minishift image list` shows number of cached images
	maxDuration, err := time.ParseDuration(maximumTime)
	if err != nil {
		return err
	}

	timeout := time.NewTimer(maxDuration)

outerPollActive:
	for {
		select {
		case <-timeout.C:
			return errors.New("Timed out in getting the number of default cached images")
		default:
			cmdOut, _, _, err := m.runner.RunCommand("image list")
			if err != nil {
				return err
			}
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
	cmdOut, _, _, err := m.runner.RunCommand("image list")
	if err != nil {
		return err
	}

	return util.CompareExpectedWithActualMatchesRegex(image, strings.TrimRight(cmdOut, "\n"))
}

func (m *Minishift) getOpenShiftInstanceUrl() string {
	cmdOut, _, _, _ := m.runner.RunCommand("ip")
	ip := strings.TrimRight(cmdOut, "\n")
	url := "https://" + ip + ":8443"
	return url
}

func (m *Minishift) getRoute(serviceName, nameSpace string) string {
	cmdOut, _, _, _ := m.runner.RunCommand("openshift service " + serviceName + " -n" + nameSpace + " --url")
	return strings.TrimRight(cmdOut, "\n")
}

func (m *Minishift) checkServiceRolloutForSuccess(service string, maxTime string, done chan rolloutMessage) {
	command := fmt.Sprintf("rollout status deploymentconfig %s --watch", service)
	ocRunner := m.runner.GetOcRunner()
	cmdOut, cmdErr, cmdExitCode, err := ocRunner.RunCommandWithTimeout(command, maxTime)
	if err != nil {
		done <- rolloutMessage{stdOut: cmdOut, stdErr: cmdErr, exitCode: cmdExitCode, err: err}
	}
	m.mutex.Lock()
	commandOutputs = append(commandOutputs,
		CommandOutput{
			command,
			cmdOut,
			cmdErr,
			cmdExitCode,
		})
	m.mutex.Unlock()

	expected := "successfully rolled out"
	// if - else construct needed, else false is returned on the second time called
	if strings.Contains(cmdOut, expected) {
		done <- rolloutMessage{stdOut: cmdOut, stdErr: cmdErr, exitCode: cmdExitCode, err: nil}
	} else {
		// get application's build logs if rollout fails
		command = fmt.Sprintf("logs bc/%s", service)
		m.ExecutingOcCommand(command)

		lastCmdResult := GetLastCommandOutput()

		cmdOut += fmt.Sprintf("\n Service build output logs: %s\n", lastCmdResult.StdOut)
		cmdErr += fmt.Sprintf("\n Service build error logs: %s\n", lastCmdResult.StdErr)
		err = fmt.Errorf("Service %v was not rolled out successfully", service)
		done <- rolloutMessage{stdOut: cmdOut, stdErr: cmdErr, exitCode: cmdExitCode, err: err}
	}
}

func (m *Minishift) rolloutServicesSuccessfully(servicesToCheck string) error {
	return m.rolloutServicesSuccessfullyBeforeTimeout(servicesToCheck, "0s")
}

func (m *Minishift) rolloutServicesSuccessfullyBeforeTimeout(servicesToCheck string, timeout string) error {
	var stdErrs []string
	var stdOuts []string
	var exitCodes []int
	var errs []error

	servicesStr := strings.Replace(servicesToCheck, ", ", " ", -1)
	servicesStr = strings.Replace(servicesStr, ",", " ", -1)
	services := strings.Split(servicesStr, " ")
	total := len(services)
	done := make(chan rolloutMessage, total)

	for i := 0; i < total; i++ {
		go m.checkServiceRolloutForSuccess(services[i], timeout, done)
	}

	success := true
	for i := 0; i < total; i++ {
		message := <-done
		stdOuts = append(stdOuts, message.stdOut)
		stdErrs = append(stdErrs, message.stdErr)
		exitCodes = append(exitCodes, message.exitCode)
		errs = append(errs, message.err)
		if message.err != nil {
			success = false
		}
		if message.exitCode != 0 {
			success = false
		}
	}

	if success != true {
		errorMessage := "Not all services successfully rolled out:\n"
		for i := 0; i < total; i++ {
			errorMessage += fmt.Sprintf("Service: '%v'\n-StdOut: %v-StdErr: %v\n-exitCode: %v\nerror:%v", services[i], stdOuts[i], stdErrs[i], exitCodes[i], errs[i])
		}
		return fmt.Errorf("Not all services successfully rolled out:\n%v", errorMessage)
	}

	return nil
}

func (m *Minishift) hostFolderMountStatus(shareName string, partern string) error {
	var mountString string
	listOfHostFolders := m.runner.GetHostfolderList()
	switch partern {
	case "should":
		mountString = shareName + "\\s*sshfs\\s*(.*?)\\s+(.*?)\\s+Y"
	case "should not":
		mountString = shareName + "\\s*sshfs\\s*(.*?)\\s+(.*?)\\s+N"
	}
	re := regexp.MustCompile(mountString)
	if !re.MatchString(listOfHostFolders) {
		return fmt.Errorf("Hostfolder status Actual: %s", listOfHostFolders)
	}
	return nil
}

func (m *Minishift) addHostFolder(shareType string, source string, target string, shareName string) error {
	testDir, _ := setupTestDirectory()
	sourcePath := filepath.Join(testDir, source)
	_, cmdErr, _, err := m.runner.RunCommand("hostfolder add -t " + shareType + " --source " + sourcePath + " --target " + target + " " + shareName)
	if cmdErr != "" {
		return errors.New(cmdErr)
	}
	return err
}
