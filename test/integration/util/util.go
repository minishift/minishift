// +build integration

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
	"path/filepath"
	"strings"
	"testing"
)

type MinishiftRunner struct {
	T          *testing.T
	BinaryPath string
	Args       string
}

func (m *MinishiftRunner) RunCommand(command string, checkError bool) string {
	commandArr := strings.Split(command, " ")
	path, _ := filepath.Abs(m.BinaryPath)
	cmd := exec.Command(path, commandArr...)
	stdout, err := cmd.Output()

	if checkError && err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			m.T.Fatalf("Error running command: %s %s. Output: %s", command, exitError.Stderr, stdout)
		} else {
			m.T.Fatalf("Error running command: %s %s. Output: %s", command, err, stdout)
		}
	}
	return string(stdout)
}

func (m *MinishiftRunner) Start() {
	m.RunCommand(fmt.Sprintf("start %s", m.Args), true)
}

func (m *MinishiftRunner) EnsureRunning() {
	if m.GetStatus() != "Running" {
		m.Start()
	}
	m.CheckStatus("Running")
}

func (m *MinishiftRunner) IsRunning() bool {
	return m.GetStatus() == "Running"
}

func (m *MinishiftRunner) EnsureDeleted() {
	m.RunCommand("delete", false)
	m.CheckStatus("Does Not Exist")
}

func (m *MinishiftRunner) SetEnvFromEnvCmdOutput(dockerEnvVars string) error {
	lines := strings.Split(dockerEnvVars, "\n")
	var envKey, envVal string
	seenEnvVar := false
	for _, line := range lines {
		fmt.Println(line)
		if strings.HasPrefix("export ", line)  {
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
	return strings.Trim(m.RunCommand("status", true), " \n")
}

func (m *MinishiftRunner) CheckStatus(desired string) {
	s := m.GetStatus()
	if s != desired {
		m.T.Fatalf("Machine is in the wrong state: %s, expected  %s", s, desired)
	}
}

type OcRunner struct {
	T          *testing.T
	BinaryPath string
}

func NewOcRunner(t *testing.T) *OcRunner {
	p, err := exec.LookPath("kubectl")
	if err != nil {
		t.Fatal("Couldn't find kubectl on path.")
	}
	return &OcRunner{BinaryPath: p, T: t}
}

func (k *OcRunner) RunCommandParseOutput(args []string, outputObj interface{}) error {
	// TODO implement (HF)
	return nil
}

func (k *OcRunner) RunCommand(args []string) (stdout []byte, err error) {
	// TODO implement (HF)
	return nil, err
}

