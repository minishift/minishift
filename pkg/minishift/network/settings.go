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

package network

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/minishift/minishift/pkg/util/os/atexit"
)

const (
	configureIPAddressMessage      = "-- Attempting to set network settings ..."
	configureIPAddressFailure      = "FAIL\n   not supported on this platform or hypervisor"
	configureNetworkScriptTemplate = `DEVICE={{.Device}}
IPADDR={{.IPAddress}}
NETMASK={{.Netmask}}
GATEWAY={{.Gateway}}
DNS1={{.DNS1}}
DNS2={{.DNS2}}
`
)

type NetworkSettings struct {
	Device    string
	IPAddress string
	Netmask   string
	Gateway   string
	DNS1      string
	DNS2      string
}

func printNetworkSettings(networkSettings NetworkSettings) {
	fmt.Println(configureIPAddressMessage)
	fmt.Println("   Device:     ", networkSettings.Device)
	fmt.Println("   IP Address: ", fmt.Sprintf("%s/%s", networkSettings.IPAddress, networkSettings.Netmask))
	if networkSettings.Gateway != "" {
		fmt.Println("   Gateway:    ", networkSettings.Gateway)
	}
	if networkSettings.DNS1 != "" {
		fmt.Println("   Nameservers:", fmt.Sprintf("%s %s", networkSettings.DNS1, networkSettings.DNS2))
	}
}

func fillNetworkSettingsScript(networkSettings NetworkSettings) string {
	result := &bytes.Buffer{}

	tmpl, err := template.New("networkScript").Parse(configureNetworkScriptTemplate)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating network script template: %s", err.Error()))
	}
	err = tmpl.Execute(result, networkSettings)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error executing network script template:: %s", err.Error()))
	}

	return result.String()
}
