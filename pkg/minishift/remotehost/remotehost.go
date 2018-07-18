/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package remotehost

import (
	"bytes"
	"fmt"
	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/minikube/sshutil"
	"golang.org/x/crypto/ssh"
)

func PrepareRemoteMachine(s *ssh.Client) error {
	osReleaseOut, err := detectOS(s)
	if err != nil {
		return err
	}
	osReleaseInfo, err := provision.NewOsRelease([]byte(osReleaseOut))
	if err != nil {
		return err
	}
	if osReleaseInfo.ID == "fedora" || osReleaseInfo.ID == "rhel" || osReleaseInfo.ID == "centos" {
		if err := prepareRHELVariant(s); err != nil {
			return err
		}
	}
	return nil
}

func detectOS(s *ssh.Client) (string, error) {
	session, err := s.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	cmd := "cat /etc/os-release"
	var buffer bytes.Buffer
	session.Stdout = &buffer
	err = session.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("Error running command '%s': %v", cmd, err)
	}
	return buffer.String(), nil
}

func prepareRHELVariant(s *ssh.Client) error {
	packageList := map[string]string{
		"firewalld": "firewall-cmd",
		"docker":    "docker",
		"net-tools": "netstat"}

	// Install required packages if not present.
	for pkg, cmd := range packageList {
		if err := sshutil.RunCommand(s, fmt.Sprintf("which %s || sudo yum install -y %s", cmd, pkg)); err != nil {
			return fmt.Errorf("Error installing package %s: %v", pkg, err)
		}
	}

	// 	Start the firewalld service.
	if err := sshutil.RunCommand(s, "sudo systemctl start firewalld"); err != nil {
		return fmt.Errorf("Error starting firewalld service: %s", err)
	}

	firewallCommandsToExecute := []string{
		"sudo firewall-cmd --permanent --add-port 2376/tcp --add-port 8443/tcp --add-port 80/tcp",
		"sudo firewall-cmd --info-zone minishift || sudo firewall-cmd --permanent --new-zone minishift",
		"sudo firewall-cmd --permanent --zone minishift --add-source 172.17.0.0/16",
		"sudo firewall-cmd --permanent --zone minishift --add-port 53/udp --add-port 8053/udp",
		"sudo firewall-cmd --reload",
	}

	for _, cmd := range firewallCommandsToExecute {
		if err := sshutil.RunCommand(s, cmd); err != nil {
			return fmt.Errorf("Error executing firewall command %s", cmd)
		}
	}
	return nil
}
