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
	"github.com/spf13/cobra"

	"fmt"

	"github.com/docker/machine/libmachine"

	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	dnsActions "github.com/minishift/minishift/pkg/minishift/network/dns"

	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

var (
	dnsStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Starts a DNS server",
		Long:  "Starts a DNS server",
		Run:   startDns,
	}
)

func startDns(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer api.Close()

	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		atexit.ExitWithMessage(1, nonExistentMachineError)
	}

	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}
	dockerCommander := docker.NewVmDockerCommander(sshCommander)

	ipAddress, err := host.Driver.GetIP()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP: %s", err.Error()))
	}

	routingSuffix := configCmd.GetDefaultRoutingSuffix(ipAddress)

	_, err = dnsActions.Start(dockerCommander, ipAddress, routingSuffix)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error starting the DNS server: %s", err.Error()))
	}
}

func init() {
	DnsCmd.AddCommand(dnsStartCmd)
}
