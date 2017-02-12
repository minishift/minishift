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
	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/provision/provisiontest"
	"github.com/docker/machine/libmachine/swarm"
	"testing"
	"text/template"
)

func TestMinishiftLogLevel(t *testing.T) {
	p := NewMinishiftProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	p.Provision(swarm.Options{}, auth.Options{}, engine.Options{LogLevel: "5"})
	if p.EngineOptions.LogLevel != "5" {
		t.Fatal("LogLevel should be 5")
	}
}

func TestMinishiftDefaultStorageDriver(t *testing.T) {
	p := NewMinishiftProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	p.Provision(swarm.Options{}, auth.Options{}, engine.Options{})
	if p.EngineOptions.StorageDriver != "devicemapper" {
		t.Fatal("Default storage driver should be devicemapper")
	}
}

func TestRhelImage(t *testing.T) {
	p := NewMinishiftProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	info := &provision.OsRelease{
		Name:      "Red Hat Enterprise Linux Server",
		VersionID: "7.3",
	}
	p.SetOsReleaseInfo(info)
	if !p.GetRedhatRelease() {
		t.Fatal("Provisioner should detect RHEL")
	}
}

func TestCentOSImage(t *testing.T) {
	p := NewMinishiftProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	info := &provision.OsRelease{
		Name:      "CentOS Linux",
		VersionID: "7.3",
	}
	p.SetOsReleaseInfo(info)
	if p.GetRedhatRelease() {
		t.Fatal("Provisioner should detect CentOS")
	}
}

func parseTemplate(t *testing.T, p *MinishiftProvisioner, engineConfigTemplate string) (*provision.DockerOptions, bytes.Buffer) {
	var (
		engineCfg bytes.Buffer
	)
	dockerOptions, err := p.GenerateDockerOptions(22)
	if err != nil {
		t.Fatal("Provisioner should Generate Docker Options")
	}

	parseTemplate, err := template.New("engineConfig").Parse(engineConfigTemplate)
	if err != nil {
		t.Fatal("Provisioner should Generate Docker Options")
	}

	engineConfigContext := provision.EngineConfigContext{
		DockerPort:       22,
		AuthOptions:      p.AuthOptions,
		EngineOptions:    p.EngineOptions,
		DockerOptionsDir: p.DockerOptionsDir,
	}
	parseTemplate.Execute(&engineCfg, engineConfigContext)
	return dockerOptions, engineCfg
}

func TestMinishiftProvisionerGenerateDockerOptionsForRHEL(t *testing.T) {
	p := NewMinishiftProvisioner("", &fakedriver.Driver{})
	info := &provision.OsRelease{
		Name:      "Red Hat Enterprise Linux Server",
		VersionID: "7.3",
	}
	p.SetOsReleaseInfo(info)

	dockerOptions, engineCfg := parseTemplate(t, p, engineConfigTemplateRHEL)

	if dockerOptions.EngineOptions != engineCfg.String() {
		t.Fatalf("Expected %s, Got %s", engineCfg.String(), dockerOptions.EngineOptions)
	}
}

func TestMinishiftProvisionerGenerateDockerOptionsForCentOS(t *testing.T) {
	var (
		engineCfg bytes.Buffer
	)
	p := NewMinishiftProvisioner("", &fakedriver.Driver{})
	info := &provision.OsRelease{
		Name:      "CentOS Linux",
		VersionID: "7.3",
	}
	p.SetOsReleaseInfo(info)

	dockerOptions, engineCfg := parseTemplate(t, p, engineConfigTemplateCentOS)

	if dockerOptions.EngineOptions != engineCfg.String() {
		t.Fatalf("Expected %s, Got %s", engineCfg.String(), dockerOptions.EngineOptions)
	}
}
