/*
Copyright (C) 2016 Red Hat, Inc.

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
	"github.com/docker/machine/drivers/hyperv"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/minishift/minishift/pkg/minikube/constants"
)

func createHypervHost(config MachineConfig) (drivers.Driver, error) {
	d := hyperv.NewDriver(constants.MachineName, constants.Minipath)
	d.Boot2DockerURL = config.GetISOFileURI()
	d.MemSize = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = int(config.DiskSize)
	d.SSHUser = "docker"

	if err := setDriverOptionsFromEnvironment(d); err != nil {
		return nil, err
	}

	return d, nil
}
