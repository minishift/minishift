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

package config

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
)

func TestHostfolderConfig(t *testing.T) {
	setup(t)
	defer teardown()

	hostfolderActual := HostFolder{
		Name: "Users",
		Type: "cifs",
		Options: map[string]string{
			"mountpoint": "",
			"uncpath":    "//127.0.0.1/Users",
			"username":   "joe@pillow.us",
			"password":   "am!g@4ever",
			"domain":     "DESKTOP-RHAIMSWIN",
		},
	}

	hostfolderExpectedMountpoint := fmt.Sprintf("%s/%s", HostfoldersDefaultPath, "Users")
	assertFields(t, hostfolderExpectedMountpoint, hostfolderActual.Mountpoint())

	viper.Set(HostfoldersMountPathKey, "/mnt/data")
	hostfolderExpectedMountpoint = "/mnt/data/Users"
	assertFields(t, hostfolderExpectedMountpoint, hostfolderActual.Mountpoint())

	hostfolderActual.Options["mountpoint"] = "/c/Users"
	hostfolderExpectedMountpoint = "/c/Users"
	assertFields(t, hostfolderExpectedMountpoint, hostfolderActual.Mountpoint())
}

func assertFields(t *testing.T, expected string, actual string) {
	if expected != actual {
		t.Errorf("Hostfolder expected: '%s'. Actual '%s'", expected, actual)
	}
}
