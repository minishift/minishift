/*
Copyright (C) 2018 Red Hat, Inc.

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

package platform

import (
	"fmt"
	"github.com/minishift/minishift/pkg/minishift/shell/powershell"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	GetGroupCmd = `(New-Object System.Security.Principal.SecurityIdentifier("S-1-5-32-578")).Translate([System.Security.Principal.NTAccount]).Value`
)

// Default External Virtual Switch Name
const ExternalVirtualSwitchName = "minishift-ext-switch"

var (
	posh             *powershell.PowerShell
	createSwitchCmds = []string{
		`$ErrorActionPreference = "Stop"`,
		`$switchName = if ($env:HYPERV_VIRTUAL_SWITCH) {$env:HYPERV_VIRTUAL_SWITCH} else {"%s"}`,
		`try { Get-VMSwitch -Name $switchName }`,
		`catch { $adapterName;`,
		`[array]$adapters = Get-NetAdapter -Physical | Where-Object { $_.status -eq "up" }`,
		`foreach ($adapter in $adapters) {`,
		`$adapterName = $adapter.Name`,
		`if ($adapterName -like "ethernet*") { break } }`,
		`New-VMSwitch -Name $switchName -NetAdapterName $adapterName`,
		`[Environment]::SetEnvironmentVariable("HYPERV_VIRTUAL_SWITCH", $switchName, "User") }`,
	}
)

func init() {
	posh = powershell.New()
}

// TODO(Refactor): Remove below code from cmd/minishift/cmd/start_preflight.go#L456
// Not doing as part of current implementation as this will have ripple effect to other code base
func checkHypervDriverUser() bool {
	posh := powershell.New()

	// Use RID to prevent issues with localized groups: https://github.com/minishift/minishift/issues/1541
	// https://support.microsoft.com/en-us/help/243330/well-known-security-identifiers-in-windows-operating-systems
	// BUILTIN\Hyper-V Administrators => S-1-5-32-578

	//Hyper-V Administrators group check fails: https://github.com/minishift/minishift/issues/2047
	//Using SecurityIdentifier overload of IsInRole()
	checkIfMemberOfHyperVAdmins :=
		`$sid = New-Object System.Security.Principal.SecurityIdentifier("S-1-5-32-578")
	@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole($sid)`
	stdOut, _, _ := posh.Execute(checkIfMemberOfHyperVAdmins)
	if !strings.Contains(stdOut, "True") {
		return false
	}

	return true
}

func addUserToHyperVAdminGroupCmds() (string, error) {
	out, _, err := posh.Execute(GetGroupCmd)
	if err != nil {
		return "", err
	}

	groupName := strings.TrimSpace(strings.Replace(strings.TrimSpace(out), "BUILTIN\\", "", -1))
	// cmd to add user to HyperV admin group
	return fmt.Sprintf("net localgroup \"%s\" \"$env:USERDOMAIN\\$env:USERNAME\" /add", groupName), nil
}

func createExternalVirtualSwitch() error {
	var (
		userToAdminCmdsContent string
		err                    error
	)

	if !checkHypervDriverUser() {
		userToAdminCmdsContent, err = addUserToHyperVAdminGroupCmds()
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Current user already in the HyperV Administrator Group ...")
	}

	createSwitchScript := fmt.Sprintf(strings.Join(createSwitchCmds, "\n"), ExternalVirtualSwitchName)

	tempDir, _ := ioutil.TempDir("", "psScripts")
	psFile, err := os.Create(filepath.Join(tempDir, "createSwitch.ps1"))
	if err != nil {
		return err
	}

	scriptContent := fmt.Sprintf("%s\n\n%s", userToAdminCmdsContent, createSwitchScript)
	psFile.WriteString(scriptContent)
	psFile.Close()

	_, _, err = posh.ExecuteAsAdmin(psFile.Name())
	if err != nil {
		return err
	}

	fmt.Printf("External swtich '%s' has been created successfully ...\n", ExternalVirtualSwitchName)
	return nil
}

func ConfigureHyperV() error {
	if err := createExternalVirtualSwitch(); err != nil {
		return err
	}

	return nil
}
