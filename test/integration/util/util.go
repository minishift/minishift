// +build integration

/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package util

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/minishift/minishift/pkg/minikube/constants"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	utilCmd "github.com/minishift/minishift/pkg/util/cmd"
)

type MinishiftRunner struct {
	CommandPath string
	CommandArgs string
}

type OcRunner struct {
	CommandPath string
}

//runCommand runs oc command with default timeout of 1 hour
func runCommand(command string, commandPath string) (stdOut string, stdErr string, exitCode int, err error) {
	return runCommandWithTimeout(command, commandPath, "1h")
}

func runCommandWithTimeout(command string, commandPath string, maxTime string) (stdOut string, stdErr string, exitCode int, err error) {
	maxDuration, err := time.ParseDuration(maxTime)
	if err != nil {
		return
	}

	var ctx context.Context
	var cancel context.CancelFunc

	commandArr := utilCmd.SplitCmdString(command)
	path, err := filepath.Abs(commandPath)
	if err != nil {
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), maxDuration)

	defer cancel()

	var outbuf, errbuf bytes.Buffer
	cmd := exec.CommandContext(ctx, path, commandArr...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	LogMessage("command", fmt.Sprintf("%s %s", commandPath, command))
	cmdErr := cmd.Run()
	stdOut = outbuf.String()
	stdErr = errbuf.String()
	LogMessage("stdout", stdOut)
	LogMessage("stderr", stdErr)

	// case 1: command timed out
	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("Command exceeded the timeout of %v.\n", maxTime)
		return
	}

	// case 2: command executed successfully
	if cmdErr == nil {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
		return
	}

	// case 3: cmd.Run() returned error - try to get exitCode
	if exitError, ok := cmdErr.(*exec.ExitError); ok {
		ws := exitError.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	} else {
		exitCode = 1
	}

	return
}

func (m *MinishiftRunner) RunCommand(command string) (stdOut string, stdErr string, exitCode int, err error) {
	stdOut, stdErr, exitCode, err = runCommand(command, m.CommandPath)
	return
}

func (m *MinishiftRunner) RunCommandAndPrintError(command string) (stdOut string, stdErr string, exitCode int, err error) {
	stdOut, stdErr, exitCode, err = runCommand(command, m.CommandPath)
	if exitCode != 0 {
		fmt.Printf("Command 'minishift %v' returned non-zero exit code: %v, StdOut: %v, StdErr: %v", command, exitCode, stdOut, stdErr)
	}
	return
}

func (m *MinishiftRunner) Start() {
	m.RunCommand(fmt.Sprintf("start %s", m.CommandArgs))
}

func (m *MinishiftRunner) CDKSetup() {
	if (os.Getenv("MINISHIFT_USERNAME") == "") || (os.Getenv("MINISHIFT_PASSWORD") == "") {
		fmt.Println("Either MINISHIFT_USERNAME or MINISHIFT_PASSWORD is not set as environment variable")
		os.Exit(1)
	}

	m.RunCommandAndPrintError(fmt.Sprintf("setup-cdk --force --minishift-home %s", os.Getenv(constants.MiniShiftHomeEnv)))
}

func (m *MinishiftRunner) IsCDK() bool {
	cmdOut, _, _, _ := m.RunCommand("setup-cdk -h")
	return strings.Contains(cmdOut, "minishift setup-cdk [flags]")
}

func (m *MinishiftRunner) IsMinishiftRunning() bool {
	return strings.Contains(m.GetStatus(), "Minishift:  Running")
}

func (m *MinishiftRunner) IsOpenshiftRunning() bool {
	return strings.Contains(m.GetStatus(), "OpenShift:  Running")
}

func (m *MinishiftRunner) GetOcRunner() *OcRunner {
	if m.IsMinishiftRunning() {
		return NewOcRunner()
	}
	return nil
}

//EnsureAllMinishiftHomesDeleted retrieves all Minishift homes which must be in format of '.<name>'
//and then deletes all VMs on all profiles on them
func (m *MinishiftRunner) EnsureAllMinishiftHomesDeleted(testDir string) {
	files, err := ioutil.ReadDir(testDir)
	if err != nil {
		fmt.Printf("Error getting files in test directory: %v\n", err)
	}

	for _, file := range files {
		if file.IsDir() == false {
			continue
		}

		dirPath := filepath.Join(testDir, file.Name())
		err = os.Setenv(constants.MiniShiftHomeEnv, dirPath)
		if err != nil {
			fmt.Printf("Error setting up environmental variable %v: %v to delete running instances.\n", constants.MiniShiftHomeEnv, err)
		}

		m.EnsureAllProfilesDeleted()
	}
}

//EnsureAllProfilesDeleted retrieves all available profiles and deletes all running VMs on them
func (m *MinishiftRunner) EnsureAllProfilesDeleted() {
	stdOut, _, _, _ := m.RunCommandAndPrintError("profile list")
	lines := strings.Split(stdOut, "\n")

	for _, line := range lines {
		re := regexp.MustCompile("- ([.\\S]+)\\s+.+")
		match := re.FindStringSubmatch(line)
		if len(match) != 2 {
			continue
		}

		profile := match[1]
		m.RunCommandAndPrintError(fmt.Sprintf("profile set %v", profile))
		err := m.EnsureDeleted()
		if err != nil {
			fmt.Printf("Error deleting profile '%v': %v", profile, err)
		}
	}
}

//EnsureDeleted deletes VM instance on currently selected profile
func (m *MinishiftRunner) EnsureDeleted() error {
	m.RunCommandAndPrintError("delete --force")

	deleted := m.CheckStatus("Does Not Exist")
	if deleted == false {
		return errors.New("Deletion of minishift instance was not successful!")
	}

	return nil
}

func (m *MinishiftRunner) SetEnvFromEnvCmdOutput(dockerEnvVars string) error {
	lines := strings.Split(dockerEnvVars, "\n")
	var envKey, envVal string
	seenEnvVar := false
	for _, line := range lines {
		fmt.Println(line)
		if strings.HasPrefix("export ", line) {
			line = strings.TrimPrefix(line, "export ")
		}
		if _, err := fmt.Sscanf(line, "export %s=\"%s\"", &envKey, &envVal); err != nil {
			seenEnvVar = true
			fmt.Println(fmt.Sprintf("%s=%s", envKey, envVal))
			os.Setenv(envKey, envVal)
		}
	}
	if seenEnvVar == false {
		return fmt.Errorf("Error: No environment variables were found in docker-env command output: %s", dockerEnvVars)
	}
	return nil
}

func (m *MinishiftRunner) GetStatus() string {
	cmdOut, _, _, _ := m.RunCommand("status")
	return strings.Trim(cmdOut, " \n")
}

func (m *MinishiftRunner) CheckStatus(desired string) bool {
	return strings.Contains(m.GetStatus(), desired)
}

func (m *MinishiftRunner) GetProfileStatus(profileName string) string {
	cmdOut, _, _, _ := m.RunCommand("--profile " + profileName + " status")
	return strings.Trim(cmdOut, " \n")
}

func (m *MinishiftRunner) GetProfileList() string {
	cmdOut, _, _, _ := m.RunCommand("profile list")
	return strings.Trim(cmdOut, " \n")
}

func (m *MinishiftRunner) GetContainerID(containerName string) (string, error) {
	cmd := fmt.Sprintf(`ssh -- 'docker ps -a -f name=%v --format {{.ID}}'`, containerName)
	cmdOut, cmdErr, _, err := m.RunCommand(cmd)
	if err != nil {
		return "", err
	}
	if cmdErr != "" {
		return "", fmt.Errorf(cmdErr)
	}

	ID := strings.Trim(cmdOut, "\n")
	if ID == "" {
		cmdOut, _, _, _ = m.RunCommand("ssh -- docker ps -a")
		fmt.Errorf("unable to get ID of container %v, command '%v' gave an empty result. Containers running:'%v'", containerName, cmd, cmdOut)
	}

	return ID, nil
}

func (m *MinishiftRunner) GetContainerStatusUsingImageID(imageID string) (string, error) {
	cmd := fmt.Sprintf(`ssh -- 'docker inspect -f {{.State.Status}} %v'`, imageID)
	cmdOut, cmdErr, _, err := m.RunCommand(cmd)
	if err != nil {
		return "", err
	}
	if cmdErr != "" {
		return strings.Trim(cmdOut, " \n"), fmt.Errorf(cmdErr)
	}

	return strings.Trim(cmdOut, " \n"), nil
}

func (m *MinishiftRunner) CpuInfo() (string, error) {
	cmdOut, cmdErr, _, _ := m.RunCommand("ssh -- cat /proc/cpuinfo")
	if cmdErr != "" {
		return strings.Trim(cmdOut, " \n"), fmt.Errorf(cmdErr)
	}
	return strings.Trim(cmdOut, " \n"), nil
}

func (m *MinishiftRunner) DiskInfo() (string, error) {
	cmdOut, cmdErr, _, _ := m.RunCommand("ssh -- sudo fdisk -l | grep Disk")
	if cmdErr != "" {
		return strings.Trim(cmdOut, " \n"), fmt.Errorf(cmdErr)
	}
	return strings.Trim(cmdOut, " \n"), nil
}

func NewOcRunner() *OcRunner {
	jsonDataPath := filepath.Join(os.Getenv(constants.MiniShiftHomeEnv), "machines", constants.MachineName+"-state.json")
	instanceState.InstanceStateConfig, _ = instanceState.NewInstanceStateConfig(jsonDataPath)
	p := instanceState.InstanceStateConfig.OcPath
	return &OcRunner{CommandPath: p}
}

// RunCommand executes oc command with default timeout of 3600s and returns standard output, error and exitcode.
func (k *OcRunner) RunCommand(command string) (stdOut string, stdErr string, exitCode int, err error) {
	stdOut, stdErr, exitCode, err = runCommand(command, k.CommandPath)
	return
}

// RunCommandWithTimeout executes oc command with timeout specified in seconds and returns standard output, error and exitcode.
func (k *OcRunner) RunCommandWithTimeout(command string, timeout string) (stdOut string, stdErr string, exitCode int, err error) {
	stdOut, stdErr, exitCode, err = runCommandWithTimeout(command, k.CommandPath, timeout)
	return
}

func (m *MinishiftRunner) GetHostfolderList() string {
	cmdOut, _, _, _ := m.RunCommand("hostfolder list")
	return strings.Trim(cmdOut, " \n")
}
