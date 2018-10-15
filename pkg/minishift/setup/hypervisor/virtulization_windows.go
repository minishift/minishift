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

package hypervisor

import (
	"errors"
	"fmt"
	"github.com/minishift/minishift/pkg/minishift/setup/platform"
	"github.com/minishift/minishift/pkg/minishift/shell/powershell"
	"github.com/minishift/minishift/pkg/util/strings"
	"regexp"
)

const (
	HypervisorCheck      = "wmic os get caption"
	WinRegex             = `Windows\s(10|8(\.1)?)\s(Home)?`
	HypervisorEnabled    = "net start"
	HypervComputeService = "Hyper-V Host Compute Service"
	HypervVMMgmt         = "Hyper-V Virtual Machine Management"
)

var (
	posh *powershell.PowerShell
)

func init() {
	posh = powershell.New()
}

func execCmd(cmd string) (stdOut string, stdErr string, err error) {
	return posh.Execute(cmd)
}

func CheckHypervisorAvailable() error {
	out, _, err := execCmd(HypervisorCheck)
	if err != nil {
		return err
	}

	if out != "" {
		match, _ := regexp.MatchString(WinRegex, out)
		if match {
			return nil
		} else {
			return errors.New(fmt.Sprintf("Hypervisor match couldn't available in \n%s", out))
		}
	}

	return nil
}

// checkHyperVEnabled will check whether HyperV (default) one
// is enabled in the current system or not
func checkHyperVEnabled() error {
	out, _, err := execCmd(HypervisorEnabled)
	if err != nil {
		return err
	}

	if out != "" {
		outArr, err := strings.SplitAndTrim(out, "\n")
		if err != nil {
			return err
		}

		if strings.Contains(outArr, HypervComputeService) && strings.Contains(outArr, HypervVMMgmt) {
			return nil
		} else {
			return errors.New(fmt.Sprintf("Hypervisor not enabled"))
		}
	}

	return nil
}

func CheckAndConfigureHypervisor() error {
	// check default Hypervisor and configure alternative if not enabled
	err := checkHyperVEnabled()
	if err != nil {
		if err := configureVirtualBox(); err != nil {
			return err
		}
	} else {
		if err := platform.ConfigureHyperV(); err != nil {
			return err
		}
	}

	return nil
}
