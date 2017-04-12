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
	"github.com/docker/machine/drivers/vmwarefusion"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/minishift/minishift/pkg/minikube/constants"
)

func createVMwareFusionHost(config MachineConfig) (drivers.Driver, error) {
	d := vmwarefusion.NewDriver(constants.MachineName, constants.Minipath).(*vmwarefusion.Driver)
	d.Boot2DockerURL = config.GetISOFileURI()
	d.Memory = config.Memory
	d.CPU = config.CPUs

	// TODO(philips): push these defaults upstream to fixup this driver
	d.SSHPort = 22
	d.ISO = d.ResolveStorePath("boot2docker.iso")

	if err := setDriverOptionsFromEnvironment(d); err != nil {
		return nil, err
	}

	return d, nil
}

type xhyveDriver struct {
	*drivers.BaseDriver
	Boot2DockerURL string
	BootCmd        string
	CPU            int
	CaCertPath     string
	DiskSize       int64
	MacAddr        string
	Memory         int
	PrivateKeyPath string
	UUID           string
	NFSShare       bool
	DiskNumber     int
	Virtio9p       bool
	Virtio9pFolder string
}

func createXhyveHost(config MachineConfig) (*xhyveDriver, error) {
	d := &xhyveDriver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: constants.MachineName,
			StorePath:   constants.Minipath,
		},
		Memory:         config.Memory,
		CPU:            config.CPUs,
		Boot2DockerURL: config.GetISOFileURI(),
		DiskSize:       int64(config.DiskSize),
		Virtio9p:       true,
		Virtio9pFolder: "/Users",
		UUID:           "F4BB3F79-AB4E-4708-95CA-E32FBFCDEFD3",
	}

	return d, nil
}
