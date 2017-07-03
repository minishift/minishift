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

package version

import (
	"strings"

	"github.com/blang/semver"
)

const VersionPrefix = "v"

// The following variables are private fields and should be set when compiling with for example --ldflags="-X github.com/minishift/minishift/pkg/version.openshiftVersion=vX.Y.Z
var (
	// The current version of minishift
	minishiftVersion = "0.0.0-unset"

	// The default version of OpenShift
	openshiftVersion = "0.0.0-unset"

	// The default version of the B2D ISO version
	b2dIsoVersion = "0.0.0-unset"

	// The default version of the CentOS ISO version
	centOsIsoVersion = "0.0.0-unset"

	// The SHA-1 of the commit this binary is build off
	commitSha = "sha-unset"
)

func GetMinishiftVersion() string {
	return minishiftVersion
}

func GetSemverVersion() (semver.Version, error) {
	return semver.Make(strings.TrimPrefix(GetMinishiftVersion(), VersionPrefix))
}

func GetOpenShiftVersion() string {
	return openshiftVersion
}

func GetB2dIsoVersion() string {
	return b2dIsoVersion
}

func GetCentOsIsoVersion() string {
	return centOsIsoVersion
}

func GetCommitSha() string {
	return commitSha
}
