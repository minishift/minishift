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

package network

import (
	"errors"
	"net"

	"github.com/docker/machine/libmachine/drivers"
)

// This will return the address as used by libmachine
func GetIP(driver drivers.Driver) (string, error) {
	ip, err := driver.GetIP()
	if err != nil {
		return "", err
	}
	return ip, nil
}

func DetermineHostIP(driver drivers.Driver) (string, error) {
	instanceIP, err := driver.GetIP()
	if err != nil {
		return "", err
	}

	for _, hostaddr := range HostIPs() {

		if NetworkContains(hostaddr, instanceIP) {
			hostip, _, _ := net.ParseCIDR(hostaddr)
			if IsIPReachable(driver, hostip.String(), false) {
				return hostip.String(), nil
			}
			return "", errors.New("unreachable")
		}
	}

	return "", errors.New("unknown error occurred")
}
