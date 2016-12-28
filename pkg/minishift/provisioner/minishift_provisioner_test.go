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
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision/provisiontest"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/libmachine/provision"
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
	if p.EngineOptions.StorageDriver != "overlay" {
		t.Fatal("Default storage driver should be overlay")
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
	if ! p.GetRedhatRelease() {
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
