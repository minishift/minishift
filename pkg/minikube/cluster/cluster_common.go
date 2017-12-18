/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package cluster

import (
	"github.com/docker/machine/drivers/generic"
	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/minishift/minishift/pkg/minikube/constants"
)

type genericDriverOptions struct {
	remoteIP              string
	remoteSSHUser         string
	sshKeyToConnectRemote string
}

func (g genericDriverOptions) String(key string) string {
	genericStringOptions := make(map[string]string)
	genericStringOptions["generic-ssh-user"] = g.remoteSSHUser
	genericStringOptions["generic-ip-address"] = g.remoteIP
	genericStringOptions["generic-ssh-key"] = g.sshKeyToConnectRemote
	return genericStringOptions[key]
}

func (g genericDriverOptions) StringSlice(key string) []string {
	return nil
}

func (g genericDriverOptions) Int(key string) int {
	genericIntOptions := make(map[string]int)
	genericIntOptions["generic-engine-port"] = 2376
	genericIntOptions["generic-ssh-port"] = 22
	return genericIntOptions[key]
}

func (g genericDriverOptions) Bool(key string) bool {
	return false
}

func createGenericDriverConfig(config MachineConfig) drivers.Driver {
	d := generic.NewDriver(constants.MachineName, constants.Minipath)
	remoteOptions := genericDriverOptions{
		remoteIP:              config.RemoteIPAddress,
		remoteSSHUser:         config.RemoteSSHUser,
		sshKeyToConnectRemote: config.SSHKeyToConnectRemote,
	}
	d.SetConfigFromFlags(remoteOptions)
	return d
}

func createVirtualboxHost(config MachineConfig) drivers.Driver {
	d := virtualbox.NewDriver(constants.MachineName, constants.Minipath)
	d.Boot2DockerURL = config.GetISOFileURI()
	d.Memory = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = int(config.DiskSize)
	d.HostOnlyCIDR = config.HostOnlyCIDR
	return d
}
