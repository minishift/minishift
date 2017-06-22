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
	"text/template"
	"time"

	"github.com/minishift/minishift/pkg/util"

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

	t.Execute(&engineCfg, engineConfigContext)

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

	log.Debugf("setting hostname %q", p.Driver.GetMachineName())
	if err := p.SetHostname(p.Driver.GetMachineName()); err != nil {
		return err
	}

	p.AuthOptions = setRemoteAuthOptions(p)
	log.Debugf("set auth options %+v", p.AuthOptions)

	log.Debugf("setting up certificates")

	configureAuth := func() error {
		if err := configureAuth(p); err != nil {
			return &util.RetriableError{Err: err}
		}
		return nil
	}

	err := util.RetryAfter(5, configureAuth, time.Second*10)
	if err != nil {
		log.Debugf("Error configuring auth during provisioning %v", err)
		return err
	}

	return nil
}
