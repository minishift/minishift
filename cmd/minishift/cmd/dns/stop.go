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

	dnsActions "github.com/minishift/minishift/pkg/minishift/network/dns"

	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

var (
	dnsStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stops the DNS server",
		Long:  "Stops the DNS server",
		Run:   stopDns,
	}
)

func stopDns(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer api.Close()

	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		atexit.ExitWithMessage(1, nonExistentMachineError)
	}

	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}
	dockerCommander := docker.NewVmDockerCommander(sshCommander)

	_, err = dnsActions.Stop(dockerCommander)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error starting the DNS server: %s", err.Error()))
	}
}

func init() {
	DnsCmd.AddCommand(dnsStopCmd)
}
