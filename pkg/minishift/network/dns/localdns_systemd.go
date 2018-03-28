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
	"strings"

	"github.com/docker/machine/libmachine/provision"

	"github.com/minishift/minishift/pkg/minishift/systemd"
)

const (
	dnsmasqServiceName = "dnsmasq"
)

var (
	dnsmasqServicePrerequisites = `sudo rm -rf /etc/dnsmasq.* /etc/resolv.dnsmasq.conf; \
sudo ln -s /var/lib/minishift/dnsmasq.hosts /etc/dnsmasq.hosts; \
sudo ln -s /var/lib/minishift/dnsmasq.conf /etc/dnsmasq.conf; \
sudo ln -s /var/lib/minishift/resolv.dnsmasq.conf /etc/resolv.dnsmasq.conf; \
sudo semanage permissive -a dnsmasq_t
`
)

type SystemdDnsService struct {
	commander *systemd.VmSystemdCommander
}

func newSystemdDnsService(sshCommander provision.SSHCommander) *SystemdDnsService {
	return &SystemdDnsService{
		commander: systemd.NewVmSystemdCommander(sshCommander),
	}
}

// isServiceRunning checks whether the dnsmasq service is running
func (s SystemdDnsService) Status() bool {
	status, err := s.commander.Status(dnsmasqServiceName)
	if err != nil || !strings.Contains(status, "active (running)") {
		return false
	}

	return true
}

func (s SystemdDnsService) Start() (bool, error) {
	cmd := fmt.Sprintf(dnsmasqServicePrerequisites)
	_, err := s.commander.Exec(cmd)
	if err != nil {
		return false, err
	}

	return s.commander.Start(dnsmasqServiceName)
}

func (s SystemdDnsService) Stop() (bool, error) {
	return s.commander.Stop(dnsmasqServiceName)
}

func (s SystemdDnsService) Restart() (bool, error) {
	return s.commander.Restart(dnsmasqServiceName)
}

func (s SystemdDnsService) Reset() {

}
