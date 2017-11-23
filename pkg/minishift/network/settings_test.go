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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillNetworkSettingsScript(t *testing.T) {
	networkSettings := NetworkSettings{
		Device:    "eth0",
		IPAddress: "10.0.75.128",
		Netmask:   "24",
		Gateway:   "10.0.75.1",
		DNS1:      "8.8.8.8",
		DNS2:      "8.8.4.4",
	}

	actual := fillNetworkSettingsScript(networkSettings)

	expected := `DEVICE=eth0
IPADDR=10.0.75.128
NETMASK=24
GATEWAY=10.0.75.1
DNS1=8.8.8.8
DNS2=8.8.4.4
`

	assert.Equal(t, expected, actual)
}
