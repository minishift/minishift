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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/provision/provisiontest"
	"github.com/docker/machine/libmachine/swarm"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/stretchr/testify/assert"
)

func TestMinishiftLogLevel(t *testing.T) {
	p := NewMinishiftProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	p.Provision(swarm.Options{}, auth.Options{}, engine.Options{LogLevel: "5"})
	assert.Equal(t, "5", p.EngineOptions.LogLevel, "Log level should be 5")
}

func TestMinishiftDefaultStorageDriver(t *testing.T) {
	p := NewMinishiftProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	p.Provision(swarm.Options{}, auth.Options{}, engine.Options{})
	assert.Equal(t, "devicemapper", p.EngineOptions.StorageDriver, "Default storage driver should be devicemapper")
}

func TestRhelImage(t *testing.T) {
	testDir := setup(t)
	defer os.RemoveAll(testDir)
	p := NewMinishiftProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	info := &provision.OsRelease{
		Name:      "Red Hat Enterprise Linux Server",
		VersionID: "7.3",
	}
	p.SetOsReleaseInfo(info)
	assert.True(t, p.GetRedhatRelease(), "Provisioner should detect RHEL")
}

func TestCentOSImage(t *testing.T) {
	p := NewMinishiftProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	info := &provision.OsRelease{
		Name:      "CentOS Linux",
		VersionID: "7.3",
	}
	p.SetOsReleaseInfo(info)
	os, _ := p.GetOsReleaseInfo()
	assert.Contains(t, os.Name, "CentOS Linux", "Provisioner should detect 'CentOS'")
}

func parseTemplate(t *testing.T, p *MinishiftProvisioner, engineConfigTemplate string) (*provision.DockerOptions, bytes.Buffer) {
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

func TestMinishiftProvisionerGenerateDockerOptionsForRHEL(t *testing.T) {
	p := NewMinishiftProvisioner("", &fakedriver.Driver{})
	info := &provision.OsRelease{
		Name:      "Red Hat Enterprise Linux Server",
		VersionID: "7.3",
	}
	p.SetOsReleaseInfo(info)

	dockerOptions, engineCfg := parseTemplate(t, p, engineConfigTemplateRHEL)

	assert.Equal(t, engineCfg.String(), dockerOptions.EngineOptions)
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

	assert.Equal(t, engineCfg.String(), dockerOptions.EngineOptions)
}

func setup(t *testing.T) string {
	// Make sure we create the required directories.
	testDir, err := ioutil.TempDir("", "minishift-provision")
	if err != nil {
		t.Error(err)
	}
	// Need to create since Minishift config get created in root command
	os.Mkdir(filepath.Join(testDir, "machines"), 0755)
	instanceConfigPath := filepath.Join(testDir, "machines", "test.json")
	minishiftConfig.InstanceConfig, err = minishiftConfig.NewInstanceConfig(instanceConfigPath)
	return testDir
}
