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
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	hvkvp "github.com/gbraad/go-hvkvp"
	"github.com/golang/glog"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/minishift/shell/powershell"
)

const (
	resendInterval        = 5 * time.Second
	resultSuccess         = "4096"
	networkingMessageName = "PROVISION_NETWORKING"
)

var (
	successCount = 0
)

func isUsingDefaultSwitch() bool {
	switchEnvName := "HYPERV_VIRTUAL_SWITCH"
	switchEnvValue := os.Getenv(switchEnvName)

	return switchEnvValue == "Default Switch" || switchEnvValue == ""
}

func getAddressAssignedToDefaultSwitch() string {
	posh := powershell.New()

	command := `Get-NetIPInterface -InterfaceAlias "vEthernet (Default Switch)" -AddressFamily IPv4 | Get-NetIPAddress | ForEach-Object { $_.IPAddress }`
	result, _, _ := posh.Execute(command)

	return strings.TrimSpace(result)
}

func determineNetmask() string {
	if isUsingDefaultSwitch() {
		return "255.255.255.240" // 28
	}
	return "255.255.255.0" // 24
}

func determineNameservers() []string {
	if isUsingDefaultSwitch() {
		return []string{getAddressAssignedToDefaultSwitch()}
	}

	posh := powershell.New()

	command := `Get-DnsClientServerAddress | Where-Object { $_.InterfaceAlias -like 'vEthernet*' } | ForEach-Object { $_.ServerAddresses }`
	result, _, _ := posh.Execute(command)

	nameservers := strings.Split(result, "\r\n")
	ipv4ns := []string{}
	for _, ns := range nameservers {
		if net.ParseIP(ns).To4() != nil {
			ipv4ns = append(ipv4ns, ns)
		}
	}

	return ipv4ns
}

func determineDefaultGateway(ipaddress string) string {
	if isUsingDefaultSwitch() {
		return getAddressAssignedToDefaultSwitch()
	}

	ip := net.ParseIP(ipaddress)
	ip = ip.To4()
	ip = ip.Mask(ip.DefaultMask())
	ip[3] = 1
	return ip.String()
}

func ConfigureNetworking(machineName string, networkSettings NetworkSettings) {
	// Instruct the user that this does not work for other Hypervisors on Windows
	if !minishiftConfig.IsHyperV() {
		fmt.Println(configureIPAddressMessage, configureIPAddressFailure)
		return
	}

	if networkSettings.Device == "" {
		networkSettings.Device = "eth0"
	}

	if networkSettings.Netmask == "" {
		fmt.Println("-- Determing netmask ... ")
		networkSettings.Netmask = determineNetmask()
	}

	if networkSettings.Gateway == "" {
		fmt.Println("-- Determing default gateway ... ")
		networkSettings.Gateway = determineDefaultGateway(networkSettings.IPAddress)
	}

	if networkSettings.DNS1 == "" && networkSettings.DNS2 == "" {
		fmt.Printf("-- Determing nameservers to use ... ")
		nameservers := determineNameservers()

		if len(nameservers) == 0 {
			fmt.Println("FAIL")
			fmt.Println("   Consider to configure using '--network-nameserver'")
		}

		if len(nameservers) > 0 {
			fmt.Println("OK")
			networkSettings.DNS1 = nameservers[0]
		}
		if len(nameservers) > 1 {
			networkSettings.DNS2 = nameservers[1]
		}
		if len(nameservers) > 2 {
			fmt.Println("   WARN: found more than 2 nameservers")
		}

	}

	printNetworkSettings(networkSettings)

	networkScript := fillNetworkSettingsScript(networkSettings)
	record := hvkvp.NewMachineKVPRecord(machineName,
		networkingMessageName,
		// to allow sending multiple lines in the value we encode the script
		base64.StdEncoding.EncodeToString([]byte(networkScript)))

	command := hvkvp.NewMachineKVPCommand(record)
	b := newConfigBasher(doConfigure, command)
	b.start()
}

func doConfigure(success chan bool, command string) {
	posh := powershell.New()
	result, _, _ := posh.Execute(command)

	if strings.Contains(result, resultSuccess) {
		if glog.V(5) {
			fmt.Printf("*")
		}
		if successCount > 3 {
			success <- true
		}
		successCount++
	}
}

type bashingFunc func(handler chan bool, command string)

type configbasher struct {
	interval time.Duration
	handler  chan bool
	bashing  bashingFunc
	command  string
}

func newConfigBasher(bashing bashingFunc, command string) *configbasher {
	return &configbasher{
		interval: resendInterval,
		handler:  make(chan bool),
		bashing:  bashing,
		command:  command,
	}
}

func (b *configbasher) start() {
	go func() {
		for {
			if glog.V(5) {
				fmt.Printf("+")
			}
			b.bashing(b.handler, b.command)
			time.Sleep(b.interval)
		}
	}()
}

func (b *configbasher) stop() {
	b.handler <- true
}
