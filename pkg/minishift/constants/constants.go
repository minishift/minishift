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
	"github.com/minishift/minishift/pkg/minikube/constants"
	"path/filepath"
)

const (
	B2dIsoAlias            = "b2d"
	CentOsIsoAlias         = "centos"
	MinikubeIsoAlias       = "minikube"
	OpenshiftContainerName = "origin"
	OpenshiftExec          = "/usr/bin/openshift"
	OpenshiftOcExec        = "/usr/bin/oc"
	DefaultProject         = "myproject"
	DefaultUser            = "developer"
	BinaryName             = "minishift"
)

var (
	ValidIsoAliases = []string{B2dIsoAlias, CentOsIsoAlias, MinikubeIsoAlias}
)

// ProfileAuthorizedKeysPath returns the path of authorized_keys file in profile dir used for authentication purpose
func ProfileAuthorizedKeysPath() string {
	return filepath.Join(constants.Minipath, "certs", "authorized_keys")
}

// ProfilePrivateKeyPath returns the path of private key of VM present in profile dir which is used for authentication purpose
func ProfilePrivateKeyPath() string {
	return filepath.Join(constants.Minipath, "certs", "id_rsa")
}
