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

package constants

import (
	"path/filepath"

	"github.com/jimmidyson/minishift/pkg/version"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/util/homedir"
)

// MachineName is the name to use for the VM.
const MachineName = "minishift"

// APIServerPort is the port that the API server should listen on.
const APIServerPort = 8443

// Fix for windows
var Minipath = filepath.Join(homedir.HomeDir(), ".minishift")

// TODO: Fix for windows
// KubeconfigPath is the path to the Kubernetes client config
var KubeconfigPath = clientcmd.RecommendedHomeFile

// MinikubeContext is the kubeconfig context name used for minishift
const MinikubeContext = "minishift"

// MiniShiftEnvPrefix is the prefix for the environmental variables
const MiniShiftEnvPrefix = "MINISHIFT"

// MakeMiniPath is a utility to calculate a relative path to our directory.
func MakeMiniPath(fileName ...string) string {
	args := []string{Minipath}
	args = append(args, fileName...)
	return filepath.Join(args...)
}

// Only pass along these flags to openshift.
var LogFlags = [...]string{
	"v",
	"vmodule",
}

const (
	DefaultMemory   = 1024
	DefaultCPUS     = 1
	DefaultDiskSize = "20g"
)

var DefaultIsoUrl = "https://github.com/jimmidyson/minishift/releases/download/v" + version.GetVersion() + "/boot2docker.iso"

const (
	RemoteOpenShiftErrPath = "/var/lib/minishift/openshift.err"
	RemoteOpenShiftOutPath = "/var/lib/minishift/openshift.out"
	RemoteOpenShiftCAPath  = "/var/lib/minishift/openshift.local.config/master/ca.crt"
)

// DockerAPIVersion is the API version implemented by Docker running in the minikube VM.
const DockerAPIVersion = "1.23"
