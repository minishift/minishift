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

package powershell

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"os/exec"
)

var (
	runAsCmds = []string{
		`$myWindowsID = [System.Security.Principal.WindowsIdentity]::GetCurrent();`,
		`$myWindowsPrincipal = New-Object System.Security.Principal.WindowsPrincipal($myWindowsID);`,
		`$adminRole = [System.Security.Principal.WindowsBuiltInRole]::Administrator;`,
		`if (-Not ($myWindowsPrincipal.IsInRole($adminRole))) {`,
		`$newProcess = New-Object System.Diagnostics.ProcessStartInfo "PowerShell";`,
		`$newProcess.Arguments = "& '" + $script:MyInvocation.MyCommand.Path + "'"`,
		`$newProcess.Verb = "runas";`,
		`[System.Diagnostics.Process]::Start($newProcess);`,
		`Exit;`,
		`}`,
	}
	isAdminCmds = []string{
		"$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())",
		"$currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)",
	}
)

type PowerShell struct {
	powerShell string
}

func New() *PowerShell {
	ps, _ := exec.LookPath("powershell.exe")
	return &PowerShell{
		powerShell: ps,
	}
}

func IsAdmin() bool {
	ps := New()
	cmd := strings.Join(isAdminCmds, ";")
	stdOut, _, err := ps.Execute(cmd)
	if err != nil {
		return false
	}
	if strings.TrimSpace(stdOut) == "False" {
		return false
	}

	return true
}

func (p *PowerShell) Execute(args ...string) (stdOut string, stdErr string, err error) {
	args = append([]string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "RemoteSigned", "-Command"}, args...)
	cmd := exec.Command(p.powerShell, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	stdOut, stdErr = stdout.String(), stderr.String()
	return
}

func (p *PowerShell) ExecuteAsAdmin(cmd string) (stdOut string, stdErr string, err error) {
	scriptContent := strings.Join(append(runAsCmds, cmd), "\n")

	tempDir, _ := ioutil.TempDir("", "psScripts")
	psFile, err := os.Create(filepath.Join(tempDir, "runAsAdmin.ps1"))
	if err != nil {
		return "", "", err
	}

	psFile.WriteString(scriptContent)
	psFile.Close()
	return p.Execute(psFile.Name())
}
