/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"path"
	"text/template"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
)

func NewBuildrootProvisioner(osReleaseID string, d drivers.Driver) *BuildrootProvisioner {
	systemdProvisioner := provision.NewSystemdProvisioner(osReleaseID, d)
	systemdProvisioner.SSHCommander = provision.GenericSSHCommander{Driver: d}
	return &BuildrootProvisioner{
		systemdProvisioner,
	}
}

type BuildrootProvisioner struct {
	provision.SystemdProvisioner
}

func (p *BuildrootProvisioner) String() string {
	return "buildroot"
}

func (p *BuildrootProvisioner) GenerateDockerOptions(dockerPort int) (*provision.DockerOptions, error) {
	var engineCfg bytes.Buffer

	driverNameLabel := fmt.Sprintf("provider=%s", p.Driver.DriverName())
	p.EngineOptions.Labels = append(p.EngineOptions.Labels, driverNameLabel)

	t, err := template.New("engineConfig").Parse(engineConfigTemplateBuildRoot)
	if err != nil {
		return nil, err
	}

	engineConfigContext := provision.EngineConfigContext{
		DockerPort:    dockerPort,
		AuthOptions:   p.AuthOptions,
		EngineOptions: p.EngineOptions,
	}

	if err := t.Execute(&engineCfg, engineConfigContext); err != nil {
		return nil, err
	}

	return &provision.DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: "/lib/systemd/system/docker.service",
	}, nil
}

func (p *BuildrootProvisioner) Package(name string, action pkgaction.PackageAction) error {
	return nil
}

func (p *BuildrootProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	p.SwarmOptions = swarmOptions
	p.AuthOptions = authOptions
	p.EngineOptions = engineOptions

	log.Info("\n   Setting hostname ... ")
	if err := p.SetHostname(p.Driver.GetMachineName()); err != nil {
		log.Info("FAIL")
		return err
	} else {
		log.Info("OK")
	}

	if err := makeDockerOptionsDir(p); err != nil {
		return err
	}

	dockerCfg, err := p.GenerateDockerOptions(engine.DefaultPort)
	if err != nil {
		return err
	}

	log.Info("Setting Docker configuration on the remote daemon...")

	if _, err = p.SSHCommand(fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s", path.Dir(dockerCfg.EngineOptionsPath), dockerCfg.EngineOptions, dockerCfg.EngineOptionsPath)); err != nil {
		return err
	}
	// This is required because for minikube ISO there is no service file for docker and GenerateDockerOptions method creates it.
	// To use the created service file in provision we need to reload daemon so that systemd able to list it.
	if _, err = p.SSHCommand("sudo systemctl -f daemon-reload"); err != nil {
		return err
	}
	// This is required because minikube ISO doesn't symlink the resolve.conf and OpenShift v3.9 have a check for this.
	// This is something we should do as part of ISO but atm not able to find a pointer [PK]
	if _, err = p.SSHCommand("sudo ln -sfn /run/systemd/resolve/resolv.conf /etc/resolv.conf"); err != nil {
		return err
	}
	p.AuthOptions = setRemoteAuthOptions(p)

	if err := provision.ConfigureAuth(p); err != nil {
		return err
	}

	doFeatureDetection(p)

	return nil
}
