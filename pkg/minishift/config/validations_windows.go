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

package config

import (
	"errors"
	"fmt"

	"github.com/minishift/minishift/pkg/minishift/shell/powershell"
)

func IsValidHypervVirtualSwitch(name string, vswitch string) error {
	posh := powershell.New()

	checkIfVirtualSwitchExists := fmt.Sprintf("Get-VMSwitch '%s'", vswitch)
	_, _, err := posh.Execute(checkIfVirtualSwitchExists)

	if err != nil {
		return errors.New("Virtual Switch not found")
	}

	return nil
}
