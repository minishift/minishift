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
	"fmt"
	"os"
	"os/exec"
)

type Runner interface {
	Run(string, ...string) error
	Output(string, ...string) ([]byte, error)
}

type RealRunner struct{}

// the real runner for the actual program, actually execs the command
func (r RealRunner) Run(command string, args ...string) error {
	fmt.Println(fmt.Sprintf("Provisioning OpenShift via '%s %s'", command, args))
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

// the real runner for get the output as byte format
func (r RealRunner) Output(command string, args ...string) ([]byte, error) {
	var (
		cmdOut []byte
		err    error
	)
	if cmdOut, err = exec.Command(command, args...).Output(); err != nil {
		return nil, err
	}
	return cmdOut, nil
}
