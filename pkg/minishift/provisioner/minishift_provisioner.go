/*
Copyright (C) 2016 Red Hat, Inc.

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

package provisioner

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
)

func NewMinishiftProvisioner(osReleaseID string, d drivers.Driver) *MinishiftProvisioner {
	systemdProvisioner := provision.NewSystemdProvisioner(osReleaseID, d)
	systemdProvisioner.SSHCommander = provision.GenericSSHCommander{Driver: d}
	return &MinishiftProvisioner{
		systemdProvisioner,
	}
}

type MinishiftProvisioner struct {
	provision.SystemdProvisioner
}

func (provisioner *MinishiftProvisioner) String() string {
	return "minishift"
}

func (provisioner *MinishiftProvisioner) SetHostname(hostname string) error {
	// we have to have SetHostname here as well to use the RedHat provisioner
	// SSHCommand to add the tty allocation
	if _, err := provisioner.SSHCommand(fmt.Sprintf(
		"sudo hostname %s && echo %q | sudo tee /etc/hostname",
		hostname,
		hostname,
	)); err != nil {
		return err
	}

	// ubuntu/debian use 127.0.1.1 for non "localhost" loopback hostnames: https://www.debian.org/doc/manuals/debian-reference/ch05.en.html#_the_hostname_resolution
	if _, err := provisioner.SSHCommand(fmt.Sprintf(
		"if grep -xq 127.0.1.1.* /etc/hosts; then sudo sed -i 's/^127.0.1.1.*/127.0.1.1 %s/g' /etc/hosts; else echo '127.0.1.1 %s' | sudo tee -a /etc/hosts; fi",
		hostname,
		hostname,
	)); err != nil {
		return err
	}

	return nil
}

func (provisioner *MinishiftProvisioner) GetRedhatRelease() bool {
	OsRelease, _ := provisioner.GetOsReleaseInfo()
	if strings.Contains(OsRelease.Name, "Red Hat Enterprise") {
		return true
	}
	return false
}

func (provisioner *MinishiftProvisioner) Package(name string, action pkgaction.PackageAction) error {
	// Since required packages is already installed, it does not support package function.
	return errors.New("Custom provisioner does not support package")
}

func (provisioner *MinishiftProvisioner) dockerDaemonResponding() bool {
	log.Debug("checking docker daemon")

	if out, err := provisioner.SSHCommand("sudo docker version"); err != nil {
		log.Warnf("Error getting SSH command to check if the daemon is running: %s", err)
		log.Debugf("'sudo docker version' output:\n%s", out)
		return false
	}

	// The daemon is up if the command worked.  Carry on.
	return true
}

func (provisioner *MinishiftProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions
	swarmOptions.Env = engineOptions.Env

	// set default storage driver for minishift
	storageDriver, err := decideStorageDriver(provisioner, "devicemapper", engineOptions.StorageDriver)
	if err != nil {
		return err
	}
	provisioner.EngineOptions.StorageDriver = storageDriver

	if err := provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	if err := mcnutils.WaitFor(provisioner.dockerDaemonResponding); err != nil {
		return err
	}

	if err := makeDockerOptionsDir(provisioner); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	if err := provision.ConfigureAuth(provisioner); err != nil {
		return err
	}

	return nil
}

func (provisioner *MinishiftProvisioner) GenerateDockerOptions(dockerPort int) (*provision.DockerOptions, error) {
	var (
		engineCfg  bytes.Buffer
		configPath = provisioner.DaemonOptionsFile
	)

	driverNameLabel := fmt.Sprintf("provider=%s", provisioner.Driver.DriverName())
	provisioner.EngineOptions.Labels = append(provisioner.EngineOptions.Labels, driverNameLabel)

	// systemd / redhat will not load options if they are on newlines
	// instead, it just continues with a different set of options; yeah...
	engineConfigTemplate := ""
	if provisioner.GetRedhatRelease() {
		engineConfigTemplate = engineConfigTemplateRHEL
	} else {
		engineConfigTemplate = engineConfigTemplateCentOS
	}

	t, err := template.New("engineConfig").Parse(engineConfigTemplate)
	if err != nil {
		return nil, err
	}

	engineConfigContext := provision.EngineConfigContext{
		DockerPort:       dockerPort,
		AuthOptions:      provisioner.AuthOptions,
		EngineOptions:    provisioner.EngineOptions,
		DockerOptionsDir: provisioner.DockerOptionsDir,
	}

	t.Execute(&engineCfg, engineConfigContext)

	daemonOptsDir := configPath
	return &provision.DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: daemonOptsDir,
	}, nil
}
