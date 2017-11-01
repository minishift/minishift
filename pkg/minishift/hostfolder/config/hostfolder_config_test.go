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

package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHostFolderConfig(t *testing.T) {
	hostFolderConfigActual := HostFolderConfig{
		Name: "Users",
		Type: "cifs",
		Options: map[string]string{
			MountPoint: "/mnt/data",
			UncPath:    "//127.0.0.1/Users",
			UserName:   "joe@pillow.us",
			Password:   "am!g@4ever",
			Domain:     "DESKTOP-RHAIMSWIN",
		},
	}

	assert.Equal(t, "/mnt/data", hostFolderConfigActual.MountPoint())
	assert.Equal(t, "joe@pillow.us", hostFolderConfigActual.Option(UserName))
	assert.Equal(t, "am!g@4ever", hostFolderConfigActual.Option(Password))
	assert.Equal(t, "DESKTOP-RHAIMSWIN", hostFolderConfigActual.Option(Domain))
}
