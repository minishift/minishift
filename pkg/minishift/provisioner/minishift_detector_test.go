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

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/minikube/tests"
	"github.com/stretchr/testify/assert"
)

func TestMinishiftProvisionerSelected(t *testing.T) {
	s, _ := tests.NewSSHServer()
	s.CommandToOutput = make(map[string]string)
	s.CommandToOutput["cat /etc/os-release"] = `NAME="CentOS Linux"
VERSION="7 (Core)"
ID="centos"
ID_LIKE="rhel fedora"
VERSION_ID="7"
PRETTY_NAME="CentOS Linux 7 (Core)"
ANSI_COLOR="0;31"
CPE_NAME="cpe:/o:centos:centos:7"
HOME_URL="https://www.centos.org/"
BUG_REPORT_URL="https://bugs.centos.org/"

CENTOS_MANTISBT_PROJECT="CentOS-7"
CENTOS_MANTISBT_PROJECT_VERSION="7"
REDHAT_SUPPORT_PRODUCT="centos"
REDHAT_SUPPORT_PRODUCT_VERSION="7"

VARIANT="minishift"
VARIANT_VERSION="1.0.0-alpha.1"
`
	port, err := s.Start()
	if err != nil {
		t.Fatalf("Error starting ssh server: %s", err)
	}
	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}

	detector := MinishiftProvisionerDetector{Delegate: provision.StandardDetector{}}
	provisioner, err := detector.DetectProvisioner(d)

	assert.NoError(t, err, "Error Getting detector")
	assert.IsType(t, new(MinishiftProvisioner), provisioner)

	osRelease, _ := provisioner.GetOsReleaseInfo()
	assert.Equal(t, "minishift", osRelease.Variant, "Release info must contain 'minishift' variant")

}

func TestDefaultCentOSProvisionerSelected(t *testing.T) {
	s, _ := tests.NewSSHServer()
	s.CommandToOutput = make(map[string]string)
	s.CommandToOutput["cat /etc/os-release"] = `NAME="CentOS Linux"
VERSION="7 (Core)"
ID="centos"
ID_LIKE="rhel fedora"
VERSION_ID="7"
PRETTY_NAME="CentOS Linux 7 (Core)"
ANSI_COLOR="0;31"
CPE_NAME="cpe:/o:centos:centos:7"
HOME_URL="https://www.centos.org/"
BUG_REPORT_URL="https://bugs.centos.org/"

CENTOS_MANTISBT_PROJECT="CentOS-7"
CENTOS_MANTISBT_PROJECT_VERSION="7"
REDHAT_SUPPORT_PRODUCT="centos"
REDHAT_SUPPORT_PRODUCT_VERSION="7"
`
	port, err := s.Start()
	assert.NoError(t, err, "Error starting ssh server")
	d := &tests.MockDriver{
		Port: port,
		BaseDriver: drivers.BaseDriver{
			IPAddress:  "127.0.0.1",
			SSHKeyPath: "",
		},
	}

	detector := MinishiftProvisionerDetector{Delegate: provision.StandardDetector{}}
	provisioner, err := detector.DetectProvisioner(d)
	assert.NoError(t, err, "Error Getting detector")

	assert.IsType(t, new(provision.CentosProvisioner), provisioner)

	osRelease, _ := provisioner.GetOsReleaseInfo()
	assert.NotEqual(t, "minishift", osRelease.Variant, "Release info should not contain 'minishift' variant")
}
