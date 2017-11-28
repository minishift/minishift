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
	"fmt"
	"net"

	"github.com/docker/machine/libmachine/drivers"
)

func IsRetrievable(driver drivers.Driver, url string, printOutput bool) bool {
	cmd := fmt.Sprintf(
		"curl -s -m 5 %s > /dev/null 2>&1",
		url)

	if printOutput {
		fmt.Printf("   Checking if '%s' is retrievable ... ", url)
	}

	if _, err := drivers.RunSSHCommandFromDriver(driver, cmd); err != nil {
		if printOutput {
			fmt.Println("FAIL")
		}
		return false
	}

	if printOutput {
		fmt.Println("OK")
	}
	return true
}

// IsIPReachable returns true is IP address is reachable from the virtual instance
func IsIPReachable(driver drivers.Driver, ip string, printOutput bool) bool {
	cmd := fmt.Sprintf(
		"sudo ping -c1 -w1 %s",
		ip)

	if printOutput {
		fmt.Printf("   Checking if '%s' is reachable ... ", ip)
	}

	if _, err := drivers.RunSSHCommandFromDriver(driver, cmd); err != nil {
		if printOutput {
			fmt.Println("FAIL")
		}
		return false
	}

	if printOutput {
		fmt.Println("OK")
	}
	return true
}

// NetworkContains returns true if the IP address belongs to the network given
func NetworkContains(network string, ip string) bool {
	_, ipnet, _ := net.ParseCIDR(network)
	address := net.ParseIP(ip)
	return ipnet.Contains(address)
}

// HostIPs returns the IP addresses assigned to the host
func HostIPs() []string {
	ips := []string{}

	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			ips = append(ips, addr.String())
		}
	}

	return ips
}
