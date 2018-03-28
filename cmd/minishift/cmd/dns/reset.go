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

	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/network/dns"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

var (
	dnsResetCmd = &cobra.Command{
		Use:   "reset",
		Short: "Resets the DNS server",
		Long:  "Resets the DNS server",
		Run:   resetDns,
	}
)

func resetDns(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer api.Close()

	host, err := api.Load(constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, nonExistentMachineError)
	}
	cmdUtil.ExitIfNotRunning(host.Driver, constants.MachineName)

	_, err = dns.Reset(host.Driver)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error starting the DNS server: %s", err.Error()))
	}
}

func init() {
	DnsCmd.AddCommand(dnsResetCmd)
}
