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

package timezone

import (
	"fmt"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
)

func SetTimeZone(host *host.Host) error {
	if minishiftConfig.InstanceStateConfig != nil {
		cmd := fmt.Sprintf("sudo timedatectl set-timezone '%s'", minishiftConfig.InstanceStateConfig.TimeZone)
		sshCommander := provision.GenericSSHCommander{Driver: host.Driver}
		if _, err := sshCommander.SSHCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func UpdateTimeZoneInConfig(timezone string) error {
	if minishiftConfig.InstanceStateConfig == nil {
		minishiftConfig.InstanceStateConfig, _ = minishiftConfig.NewInstanceStateConfig(minishiftConstants.GetInstanceStateConfigPath())
	}
	minishiftConfig.InstanceStateConfig.TimeZone = timezone
	return minishiftConfig.InstanceStateConfig.Write()
}

func GetTimeZone(host *host.Host) (string, error) {
	cmd := "sudo timedatectl status"
	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}
	output, err := sshCommander.SSHCommand(cmd)
	if err != nil {
		return "", err
	}
	return output, nil
}

func GetTimeZoneList(host *host.Host) (string, error) {
	cmd := "sudo timedatectl list-timezones"
	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}
	output, err := sshCommander.SSHCommand(cmd)
	if err != nil {
		return "", err
	}
	return output, nil
}
