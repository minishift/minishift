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

	"github.com/spf13/viper"
)

type HostFolder struct {
	Name    string
	Type    string
	Options map[string]string
}

const (
	HostfoldersDefaultPath  = "/mnt/sda1"
	HostfoldersMountPathKey = "hostfolders-mountpath"
)

func GetHostfoldersMountPath(name string) string {
	overrideMountPath := viper.GetString(HostfoldersMountPathKey)
	if len(overrideMountPath) > 0 {
		return fmt.Sprintf("%s/%s", overrideMountPath, name)
	}

	return fmt.Sprintf("%s/%s", HostfoldersDefaultPath, name)
}

func (hf *HostFolder) Mountpoint() string {
	overrideMountPoint := hf.Options["mountpoint"]
	if len(overrideMountPoint) > 0 {
		return overrideMountPoint
	}

	return GetHostfoldersMountPath(hf.Name)
}
