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

package hostfolder

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/minishift/minishift/pkg/minishift/hostfolder/config"
)

type Type int

const (
	// SSHFS defines the constant to be used for the SSFS host folder type.
	SSHFS Type = iota

	// CIFS defines the constant to be used for the CIFS host folder type.
	CIFS
)

func (t Type) String() string {
	names := [...]string{
		"sshfs",
		"cifs"}

	// prevent panicking
	if t < SSHFS || t > CIFS {
		return "unknown"
	}
	return names[t]
}

type HostFolder interface {
	// Config returns the host folder configuration for this HostFolder.
	Config() config.HostFolderConfig

	// Mount mounts the host folder specified by name into the running VM. nil is returned on success.
	// An error is returned, if the VM is not running, the specified host folder does not exist or the mount fails.
	Mount(driver drivers.Driver) error

	// Umount umounts the host folder specified by name. nil is returned on success.
	// An error is returned, if the VM is not running, the specified host folder does not exist or the mount fails.
	Umount(driver drivers.Driver) error
}
