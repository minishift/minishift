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
	"github.com/minishift/minishift/pkg/util/os"
	"os/exec"
	"strings"
)

const (
	GetGroupCmd = `(New-Object System.Security.Principal.SecurityIdentifier("S-1-5-32-578")).Translate([System.Security.Principal.NTAccount]).Value`
)

// Default External Virtual Switch Name
const ExternalVirtualSwitchName = "minishift-external"

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
func CheckHypervDriverUser() bool {
	posh := powershell.New()

	// Use RID to prevent issues with localized groups: https://github.com/minishift/minishift/issues/1541
	// https://support.microsoft.com/en-us/help/243330/well-known-security-identifiers-in-windows-operating-systems
	// BUILTIN\Hyper-V Administrators => S-1-5-32-578

	//Hyper-V Administrators group check fails: https://github.com/minishift/minishift/issues/2047
	//Using SecurityIdentifier overload of IsInRole()
	checkIfMemberOfHyperVAdmins :=
		`$sid = New-Object System.Security.Principal.SecurityIdentifier("S-1-5-32-578")
	@([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole($sid)`
	stdOut, _, _ := posh.ExecuteAsAdmin(checkIfMemberOfHyperVAdmins)
	if !strings.Contains(stdOut, "True") {
		return false
	}

	return true
}

// checkHypervDriverInstalled returns true if Hyper-V driver is installed
func checkHypervDriverInstalled() bool {
	// Check if Hyper-V's Virtual Machine Management Service is installed
	_, err := exec.LookPath("vmms.exe")
	if err != nil {
		return false
	}

	// check to see if a hypervisor is present. if hyper-v is installed and enabled,
	posh := powershell.New()

	checkHypervisorPresent := `@(Get-Wmiobject Win32_ComputerSystem).HypervisorPresent`

	stdOut, _, _ := posh.Execute(checkHypervisorPresent)
	if !strings.Contains(stdOut, "True") {
		return false
	}

	return true
}

func AddUserToHyperVAdminGroup() error {
	var (
		username string
		err      error
	)

	username, err = os.CurrentUser()
	if err != nil {
		return err
	}

	out, _, err := posh.Execute(GetGroupCmd)
	if err != nil {
		return err
	}

	groupName := strings.TrimSpace(strings.Replace(strings.TrimSpace(out), "BUILTIN\\", "", -1))
	cmd := fmt.Sprintf(`net localgroup "%s" "$env:USERDOMAIN\$env:USERNAME" /add`, groupName)
	_, _, err = posh.ExecuteAsAdmin(cmd)
	if err != nil {
		return err
	}

	fmt.Printf("User '%s' has been added to '%s' group successfully.", username, groupName)
	return nil
}

func CreateExternalVirtualSwitch() error {
	createSwitchScript := fmt.Sprintf(strings.Join(createSwitchCmds, "\n"), ExternalVirtualSwitchName)
	_, _, err := posh.ExecuteAsAdmin(createSwitchScript)
	if err != nil {
		return err
	}

	fmt.Printf("\nExternal swtich '%s' has been created successfully.\n", ExternalVirtualSwitchName)
	return nil
}

func enableHyperV() error {
	cmd := `Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All`
	fmt.Println("Enabling HyperV ...")
	_, _, err := posh.ExecuteAsAdmin(cmd)
	if err != nil {
		return err
	}

	fmt.Println("Enabled HyperV.")
	return nil
}

func EnableHyperV() error {
	var rebootConfirm string

	if checkHypervDriverInstalled() {
		return nil
	}

	if err := enableHyperV(); err != nil {
		return err
	}

	fmt.Printf("Setup needs to reboot your machine to enable the Hyper-V.\n" +
		"After reboot, continue running 'minishift setup' to finish the remaining setup? [y/N]: ")
	fmt.Scanln(&rebootConfirm)

	if strings.ToLower(rebootConfirm) == "y" {
		posh.Execute(`Restart-Computer`)
	}

	return nil
}
