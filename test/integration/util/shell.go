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

package util

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/minishift/minishift/pkg/util/os"
)

const (
	exitCodeIdentifier = "exitCodeOfLastCommandInShell="

	bashExitCodeCheck       = "echo %v$?"
	fishExitCodeCheck       = "echo %v$status"
	tcshExitCodeCheck       = "echo %v$?"
	zshExitCodeCheck        = "echo %v$?"
	cmdExitCodeCheck        = "echo %v%%errorlevel%%"
	powershellExitCodeCheck = "echo %v$lastexitcode"
)

var (
	shell shellInstance
)

type shellInstance struct {
	startArgument    []string
	name             string
	checkExitCodeCmd string

	minishiftPath string

	instance *exec.Cmd
	outbuf   bytes.Buffer
	errbuf   bytes.Buffer
	excbuf   bytes.Buffer

	outPipe io.ReadCloser
	errPipe io.ReadCloser
	inPipe  io.WriteCloser

	outScanner *bufio.Scanner
	errScanner *bufio.Scanner

	stdoutChannel   chan string
	stderrChannel   chan string
	exitCodeChannel chan string
}

func (shell shellInstance) getLastShellOutput(stdType string) string {
	var returnValue string
	switch stdType {
	case "stdout":
		returnValue = shell.outbuf.String()
	case "stderr":
		returnValue = shell.errbuf.String()
	case "exitcode":
		returnValue = shell.excbuf.String()
	default:
		fmt.Printf("Field '%s' of shell's output is not supported. Only 'stdout', 'stderr' and 'exitcode' are supported.", stdType)
	}

	return returnValue
}

func (shell *shellInstance) scanPipe(scanner *bufio.Scanner, buffer *bytes.Buffer, stdType string, channel chan string) {
	for scanner.Scan() {
		str := scanner.Text()
		LogMessage(stdType, str)
		buffer.WriteString(str + "\n")

		if strings.Contains(str, exitCodeIdentifier) && !strings.Contains(str, shell.checkExitCodeCmd) {
			exitCode := strings.Split(str, "=")[1]
			shell.exitCodeChannel <- exitCode
		}
	}

	return
}

func (shell *shellInstance) close() error {
	closingCmd := "exit\n"
	io.WriteString(shell.inPipe, closingCmd)
	err := shell.instance.Wait()
	if err != nil {
		fmt.Println("Error closing shell instance:", err)
	}

	shell.instance = nil

	return err
}

func (shell *shellInstance) configureTypeOfShell(shellName string) {
	switch shellName {
	case "bash":
		shell.name = shellName
		shell.checkExitCodeCmd = fmt.Sprintf(bashExitCodeCheck, exitCodeIdentifier)
	case "tcsh":
		shell.name = shellName
		shell.checkExitCodeCmd = fmt.Sprintf(tcshExitCodeCheck, exitCodeIdentifier)
	case "zsh":
		shell.name = shellName
		shell.checkExitCodeCmd = fmt.Sprintf(zshExitCodeCheck, exitCodeIdentifier)
	case "cmd":
		shell.name = shellName
		shell.checkExitCodeCmd = fmt.Sprintf(cmdExitCodeCheck, exitCodeIdentifier)
	case "powershell":
		shell.name = shellName
		shell.startArgument = []string{"-Command", "-"}
		shell.checkExitCodeCmd = fmt.Sprintf(powershellExitCodeCheck, exitCodeIdentifier)
	case "fish":
		shell.name = "default"
		fmt.Println("Fish shell is currently not supported by integration tests. Default shell for the OS will be used.")
		fallthrough
	default:
		if shell.name != "" {
			fmt.Printf("Shell %v is not supported, will set the default shell for the OS to be used.\n", shell.name)
		}
		switch os.CurrentOS() {
		case "darwin", "linux":
			shell.name = "bash"
			shell.checkExitCodeCmd = fmt.Sprintf(bashExitCodeCheck, exitCodeIdentifier)
		case "windows":
			shell.name = "powershell"
			shell.startArgument = []string{"-Command", "-"}
			shell.checkExitCodeCmd = fmt.Sprintf(powershellExitCodeCheck, exitCodeIdentifier)
		}
	}

	return
}

func (shell *shellInstance) start(shellName string, minishiftPath string) error {
	var err error

	shell.configureTypeOfShell(shellName)
	shell.minishiftPath = minishiftPath
	shell.stdoutChannel = make(chan string)
	shell.stderrChannel = make(chan string)
	shell.exitCodeChannel = make(chan string)

	shell.instance = exec.Command(shell.name, shell.startArgument...)

	shell.outPipe, err = shell.instance.StdoutPipe()
	if err != nil {
		return err
	}

	shell.errPipe, err = shell.instance.StderrPipe()
	if err != nil {
		return err
	}

	shell.inPipe, err = shell.instance.StdinPipe()
	if err != nil {
		return err
	}

	shell.outScanner = bufio.NewScanner(shell.outPipe)
	shell.errScanner = bufio.NewScanner(shell.errPipe)

	go shell.scanPipe(shell.outScanner, &shell.outbuf, "stdout", shell.stdoutChannel)
	go shell.scanPipe(shell.errScanner, &shell.errbuf, "stderr", shell.stderrChannel)

	err = shell.instance.Start()
	if err != nil {
		return err
	}

	fmt.Printf("The %v instance has been started and will be used for testing.\n", shell.name)
	return err
}

func StartHostShellInstance(shellName string, minishiftPath string) error {
	return shell.start(shellName, minishiftPath)
}

func CloseHostShellInstance() error {
	return shell.close()
}

func ExecuteInHostShell(command string) error {
	var err error

	if shell.instance == nil {
		return errors.New("Shell instance is started.")
	}

	shell.outbuf.Reset()
	shell.errbuf.Reset()
	shell.excbuf.Reset()

	LogMessage(shell.name, command)

	_, err = io.WriteString(shell.inPipe, command+"\n")
	if err != nil {
		return err
	}

	_, err = shell.inPipe.Write([]byte(shell.checkExitCodeCmd + "\n"))
	if err != nil {
		return err
	}

	exitCode := <-shell.exitCodeChannel
	shell.excbuf.WriteString(exitCode)

	return err
}

func ExecuteInHostShellSucceedsOrFails(command string, expectedResult string) error {
	err := ExecuteInHostShell(command)
	if err != nil {
		return err
	}

	exitCode := shell.excbuf.String()

	if expectedResult == "succeeds" && exitCode != "0" {
		err = fmt.Errorf("Command '%s', expected to succeed, exited with exit code: %s\n", command, exitCode)
	}
	if expectedResult == "fails" && exitCode == "0" {
		err = fmt.Errorf("Command '%s', expected to fail, exited with exit code: %s\n", command, exitCode)
	}

	return err
}

func ExecuteInHostShellLineByLine() error {
	var err error
	stdout := shell.getLastShellOutput("stdout")
	commandArray := strings.Split(stdout, "\n")
	for index := range commandArray {
		if !strings.Contains(commandArray[index], exitCodeIdentifier) {
			err = ExecuteInHostShell(commandArray[index])
		}
	}

	return err
}

func ExecuteMinishiftInHostShell(commandField string) error {
	command := shell.minishiftPath + " " + commandField
	return ExecuteInHostShell(command)
}

func ExecuteMinishiftInHostShellSucceedsOrFails(commandField string, expected string) error {
	command := shell.minishiftPath + " " + commandField
	return ExecuteInHostShellSucceedsOrFails(command, expected)
}

func HostShellCommandReturnShouldContain(commandField string, expected string) error {
	return CompareExpectedWithActualContains(expected, shell.getLastShellOutput(commandField))
}

func HostShellCommandReturnShouldNotContain(commandField string, notexpected string) error {
	return CompareExpectedWithActualNotContains(notexpected, shell.getLastShellOutput(commandField))
}

func HostShellCommandReturnShouldContainContent(commandField string, expected *gherkin.DocString) error {
	return CompareExpectedWithActualContains(expected.Content, shell.getLastShellOutput(commandField))
}

func HostShellCommandReturnShouldNotContainContent(commandField string, notexpected *gherkin.DocString) error {
	return CompareExpectedWithActualNotContains(notexpected.Content, shell.getLastShellOutput(commandField))
}

func HostShellCommandReturnShouldEqual(commandField string, expected string) error {
	return CompareExpectedWithActualEquals(expected, shell.getLastShellOutput(commandField))
}

func HostShellCommandReturnShouldEqualContent(commandField string, expected *gherkin.DocString) error {
	return CompareExpectedWithActualEquals(expected.Content, shell.getLastShellOutput(commandField))
}
