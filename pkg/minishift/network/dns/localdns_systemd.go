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
	"fmt"
	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/minishift/systemd"
	"strings"
)

// isServiceRunning checks whether the dnsmasq service is running
func isServiceRunning(sshCommander provision.SSHCommander) bool {
	systemdCommander := systemd.NewVmSystemdCommander(sshCommander)

	status, err := systemdCommander.Status(dnsmasqServiceName)
	if err != nil || !strings.Contains(status, "active (running)") {
		return false
	}

	return true
}

func startService(sshCommander provision.SSHCommander) (bool, error) {
	systemdCommander := systemd.NewVmSystemdCommander(sshCommander)

	cmd := fmt.Sprintf(dnsmasqServicePrerequisites)
	_, err := systemdCommander.Exec(cmd)
	if err != nil {
		return false, err
	}

	return systemdCommander.Start(dnsmasqServiceName)
}

func stopService(sshCommander provision.SSHCommander) (bool, error) {
	systemdCommander := systemd.NewVmSystemdCommander(sshCommander)

	return systemdCommander.Stop(dnsmasqServiceName)
}
