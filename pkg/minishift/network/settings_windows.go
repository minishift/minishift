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
	"strings"
	"time"

	hvkvp "github.com/gbraad/go-hvkvp"
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/minishift/shell/powershell"
)

const (
	resendInterval        = 5 * time.Second
	resultSuccess         = "4096"
	networkingMessageName = "PROVISION_NETWORKING"
)

func ConfigureNetworking(machineName string, vmDriver string, networkSettings NetworkSettings) {
	// Instruct the user that this does not work for other Hypervisors on Windows
	if vmDriver != "hyperv" {
		fmt.Println(configureIPAddressMessage, configureIPAddressFailure)
		return
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
	result, _ := posh.Execute(command)

	if strings.Contains(result, resultSuccess) {
		if glog.V(5) {
			fmt.Printf("*")
		}
		success <- true
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
