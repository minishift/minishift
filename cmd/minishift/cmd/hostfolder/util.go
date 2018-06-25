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
	"fmt"
	cmdConfig "github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/minishift/hostfolder"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/viper"
	"regexp"
	"strings"
)

const (
	HostfoldersDefaultPath  = "/mnt/sda1"
	HostfoldersMountPathKey = "hostfolders-mountpath"
)

var (
	optionsRegexp  = regexp.MustCompile(`^([a-z]+=.*,)?([a-z]+=.*)$`)
	keyValueRegexp = regexp.MustCompile(`^([a-z]+)=(.*)$`)
)

func getHostFolderManager() *hostfolder.Manager {
	hostFolderManager, err := hostfolder.NewManager(config.InstanceStateConfig, config.AllInstancesConfig)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	port := viper.GetInt(cmdConfig.HostFoldersSftpPort.Name)
	if port != 0 {
		hostfolder.SftpPort = port
	}

	return hostFolderManager
}

func readInputForMountPoint(name string) string {
	defaultMountPoint := getHostFolderMountPath(name)
	mountPointText := fmt.Sprintf("Mountpoint [%s]", defaultMountPoint)

	mountPoint := util.ReadInputFromStdin(mountPointText)
	if len(mountPoint) == 0 {
		mountPoint = defaultMountPoint
	}
	return mountPoint
}

func getHostFolderMountPath(name string) string {
	overrideMountPath := viper.GetString(HostfoldersMountPathKey)
	if len(overrideMountPath) > 0 {
		return fmt.Sprintf("%s/%s", overrideMountPath, name)
	}

	return fmt.Sprintf("%s/%s", HostfoldersDefaultPath, name)
}

func getOptions(optionString string) map[string]string {
	options := make(map[string]string)

	var key, value, remainder string
	remainder = optionString
	for remainder != "" {
		result := optionsRegexp.FindAllStringSubmatch(remainder, -1)
		for _, match := range result {
			key, value, remainder = extractMatch(match)
			options[key] = value
		}
	}

	return options
}

func extractMatch(match []string) (string, string, string) {
	result := keyValueRegexp.FindAllStringSubmatch(match[2], -1)
	remainder := strings.TrimRight(match[1], ",")
	return result[0][1], result[0][2], remainder
}
