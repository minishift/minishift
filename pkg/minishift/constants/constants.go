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

package constants

import (
	"fmt"
	"path/filepath"

	"github.com/minishift/minishift/pkg/minikube/constants"
)

const (
	B2dIsoAlias                    = "b2d"
	CentOsIsoAlias                 = "centos"
	OpenshiftContainerName         = "origin"
	OpenshiftApiContainerLabel     = "io.kubernetes.container.name=apiserver"
	KubernetesApiContainerLabel    = "io.kubernetes.container.name=api"
	OpenshiftExec                  = "/usr/bin/openshift"
	OpenshiftOcExec                = "/usr/bin/oc"
	DefaultProject                 = "myproject"
	DefaultUser                    = "developer"
	DefaultUserPassword            = "developer"
	SkipVerifyInsecureTLS          = "insecure-skip-tls-verify=true"
	BinaryName                     = "minishift"
	OcPathInsideVM                 = "/var/lib/minishift/bin"
	BaseDirInsideInstance          = "/var/lib/minishift/base"
	ImageNameForClusterUpImageFlag = "openshift/origin-${component}"
)

var (
	ValidIsoAliases = []string{B2dIsoAlias, CentOsIsoAlias}
	ValidComponents = []string{"automation-service-broker", "service-catalog", "template-service-broker"}
)

// ProfileAuthorizedKeysPath returns the path of authorized_keys file in profile dir used for authentication purpose
func ProfileAuthorizedKeysPath() string {
	return filepath.Join(constants.Minipath, "certs", "authorized_keys")
}

// ProfilePrivateKeyPath returns the path of private key of VM present in profile dir which is used for authentication purpose
func ProfilePrivateKeyPath() string {
	return filepath.Join(constants.Minipath, "certs", "id_rsa")
}

func GetOpenshiftImageToFetchOC(openshiftVersion string, isGreaterOrEqualToBaseVersion bool) string {
	if isGreaterOrEqualToBaseVersion {
		return fmt.Sprintf("openshift/origin-control-plane:%s", openshiftVersion)
	}
	return fmt.Sprintf("openshift/origin:%s", openshiftVersion)
}

// GetInstanceStateConfigPath return the path of instance config json file
func GetInstanceStateConfigPath() string {
	return filepath.Join(constants.Minipath, "machines", constants.MachineName+"-state.json")
}

// GetInstanceStateConfigOldPath return the old path of instance config to make new binary backward compatible
func GetInstanceStateConfigOldPath() string {
	return filepath.Join(constants.Minipath, "machines", constants.MachineName+".json")
}

// GetInstanceConfigPath return the path of instance config json file
func GetInstanceConfigPath() string {
	return filepath.Join(constants.Minipath, "config", constants.MachineName+".json")
}
