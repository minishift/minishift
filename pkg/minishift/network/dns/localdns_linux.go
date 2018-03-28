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

package dns

import (
	"fmt"
	"github.com/minishift/minishift/pkg/minishift/network"
)

func handleHostDNSSettingsAfterStart(ipAddress string) (bool, error) {
	fmt.Println(fmt.Sprintf("Add as first line in /etc/resolv.conf: nameserver %s", ipAddress))

	return true, nil
}

func handleHostDNSSettingsAfterStop(ipAddress string) (bool, error) {
	if has, _ := network.HasNameserverConfiguredLocally(ipAddress); has == true {
		fmt.Println(fmt.Sprintf("Please remove the entry for %s from /etc/resolv.conf", ipAddress))
	}
	return true, nil
}

func handleHostDNSSettingsAfterReset() (bool, error) {

	return true, nil
}
