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

package provisioner

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/provision/provisiontest"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/stretchr/testify/assert"
)

func TestBuildRootLogLevel(t *testing.T) {
	p := NewBuildrootProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	p.Provision(swarm.Options{}, auth.Options{}, engine.Options{LogLevel: "5"})
	assert.Equal(t, "5", p.EngineOptions.LogLevel)
}

func TestBuildRootProvisionerGenerateDockerOptions(t *testing.T) {
	p := NewBuildrootProvisioner("", &fakedriver.Driver{})

	dockerOptions, engineCfg := parseBuildRootTemplate(t, p, engineConfigTemplateBuildRoot)

	assert.Equal(t, engineCfg.String(), dockerOptions.EngineOptions)
}

func parseBuildRootTemplate(t *testing.T, p *BuildrootProvisioner, engineConfigTemplate string) (*provision.DockerOptions, bytes.Buffer) {
	var (
		engineCfg bytes.Buffer
	)
	dockerOptions, err := p.GenerateDockerOptions(22)
	assert.NoError(t, err, "Provisioner should Generate Docker Options")

	parseTemplate, err := template.New("engineConfig").Parse(engineConfigTemplate)
	assert.NoError(t, err, "Provisioner should Generate Docker Options")

	engineConfigContext := provision.EngineConfigContext{
		DockerPort:       22,
		AuthOptions:      p.AuthOptions,
		EngineOptions:    p.EngineOptions,
		DockerOptionsDir: p.DockerOptionsDir,
	}
	parseTemplate.Execute(&engineCfg, engineConfigContext)
	return dockerOptions, engineCfg
}
