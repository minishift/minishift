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
	CentOsIsoAlias                 = "centos"
	OpenshiftContainerName         = "origin"
	OpenshiftApiContainerLabel     = "io.kubernetes.container.name=apiserver"
	KubernetesApiContainerLabel    = "io.kubernetes.container.name=api"
	OpenshiftOcExec                = "/usr/bin/oc"
	DefaultProject                 = "myproject"
	DefaultUser                    = "developer"
	DefaultUserPassword            = "developer"
	BinaryName                     = "minishift"
	OcPathInsideVM                 = "/var/lib/minishift/bin"
	BaseDirInsideInstance          = "/var/lib/minishift/base"
	ImageNameForClusterUpImageFlag = "openshift/origin-${component}"
	HypervDefaultVirtualSwitchId   = "c08cb7b8-9b3c-408e-8e30-5e16a3aeb444"
	HypervDefaultVirtualSwitchName = "Default Switch"
	DockerbridgeSubnetCmd          = `docker network inspect -f "{{range .IPAM.Config }}{{ .Subnet }}{{end}}" bridge`
	MinishiftEnableExperimental    = "MINISHIFT_ENABLE_EXPERIMENTAL"
	SystemtrayDaemon               = "systemtray"
	SftpdDaemon                    = "sftpd"
	ProxyDaemon                    = "proxy"
)

var (
	ValidIsoAliases = []string{CentOsIsoAlias}
	ValidComponents = []string{"automation-service-broker", "service-catalog", "template-service-broker"}
	ValidServices   = []string{SystemtrayDaemon, SftpdDaemon, ProxyDaemon}
)

// ProfileAuthorizedKeysPath returns the path of authorized_keys file in profile dir used for authentication purpose
func ProfileAuthorizedKeysPath() string {
	return filepath.Join(constants.Minipath, "certs", "authorized_keys")
}

// ProfilePrivateKeyPath returns the path of private key of VM present in profile dir which is used for authentication purpose
func ProfilePrivateKeyPath() string {
	return filepath.Join(constants.Minipath, "certs", "id_rsa")
}

func GetOpenshiftImageToFetchOC(openshiftVersion string) string {
	return fmt.Sprintf("openshift/origin-control-plane:%s", openshiftVersion)
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

// GetProfileInstanceConfigPath return the path of instance config json file for a profile
func GetProfileInstanceConfigPath(profileName string) string {
	return filepath.Join(constants.GetProfileHomeDir(profileName), "config", profileName+".json")
}
