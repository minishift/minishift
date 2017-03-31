/*
Copyright (C) 2016 Red Hat, Inc.

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
	"io"
	"os/exec"
	"path/filepath"
	"syscall"
)

type Runner interface {
	Run(stdOut io.Writer, stdErr io.Writer, commandPath string, args ...string) int
	Output(string, ...string) ([]byte, error)
}

type RealRunner struct{}

// the real runner for get the output as byte format
func (r RealRunner) Output(command string, args ...string) ([]byte, error) {
	var (
		cmdOut []byte
		err    error
	)

	if cmdOut, err = exec.Command(command, args...).CombinedOutput(); err != nil {
		return nil, err
	}
	return cmdOut, nil
}

func (r RealRunner) Run(stdOut io.Writer, stdErr io.Writer, commandPath string, args ...string) int {
	path, _ := filepath.Abs(commandPath)
	cmd := exec.Command(path, args...)

	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	err := cmd.Run()

	var exitCode int
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			exitCode = 1 // unable to get error code
		}
	} else {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}
	return exitCode
}
