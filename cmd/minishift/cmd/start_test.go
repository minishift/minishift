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

package cmd

import (
	"testing"

	minitesting "github.com/minishift/minishift/pkg/testing"

	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minikube/tests"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util"
	"github.com/spf13/viper"
)

var (
	ocClusterUpHelp = []byte(`Starts an OpenShift cluster using Docker containers, provisioning a registry, router, initial templates, and a default
project.

This command will attempt to use an existing connection to a Docker daemon. Before running the command, ensure that you
can execure docker commands successfully (i.e. 'docker ps').

Optionally, the command can create a new Docker machine for OpenShift using the VirtualBox driver when the
--create-machine argument is specified. The machine will be named 'openshift' by default. To name the machine
differently, use the --docker-machine=NAME argument. If the --docker-machine=NAME argument is specified, but
--create-machine is not, the command will attempt to find an existing docker machine with that name and start it if it's
not running.

By default, the OpenShift cluster will be setup to use a routing suffix that ends in xip.io. This is to allow dynamic
host names to be created for routes. An alternate routing suffix can be specified using the --routing-suffix flag.

A public hostname can also be specified for the server with the --public-hostname flag.

Usage:
  oc cluster up [options]

Examples:
  # Start OpenShift on a new docker machine named 'openshift'
  oc cluster up --create-machine

  # Start OpenShift using a specific public host name
  oc cluster up --public-hostname=my.address.example.com

  # Start OpenShift and preserve data and config between restarts
  oc cluster up --host-data-dir=/mydata --use-existing-config

  # Use a different set of images
  oc cluster up --image="registry.example.com/origin" --version="v1.1"

Options:
      --create-machine=false: If true, create a Docker machine if one doesn't exist
      --docker-machine='': Specify the Docker machine to use
  -e, --env=[]: Specify a key-value pair for an environment variable to set on OpenShift container
      --forward-ports=true: If true, use Docker port-forwarding to communicate with origin container. Requires 'socat'
locally.
      --host-config-dir='/var/lib/origin/openshift.local.config': Directory on Docker host for OpenShift configuration
      --host-data-dir='': Directory on Docker host for OpenShift data. If not specified, etcd data will not be persisted
on the host.
      --host-pv-dir='/var/lib/origin/openshift.local.pv': Directory on host for OpenShift persistent volumes
      --host-volumes-dir='/var/lib/origin/openshift.local.volumes': Directory on Docker host for OpenShift volumes
      --http-proxy='': HTTP proxy to use for master and builds
      --https-proxy='': HTTPS proxy to use for master and builds
      --image='openshift/origin': Specify the images to use for OpenShift
      --image-streams='centos7': Specify which image streams to use, centos7|rhel7
      --logging=false: If true, install logging (experimental)
      --metrics=false: If true, install metrics (experimental)
      --no-proxy=[]: List of hosts or subnets for which a proxy should not be used
      --public-hostname='': Public hostname for OpenShift cluster
      --routing-suffix='': Default suffix for server routes
      --server-loglevel=0: Log level for OpenShift server
      --skip-registry-check=false: If true, skip Docker daemon registry check
      --use-existing-config=false: If true, use existing configuration if present
      --version='': Specify the tag for OpenShift images

Use "oc options" for a list of global command-line options (applies to all commands).`)
)

type RecordingRunner struct {
	Cmd  string
	Args []string
}

func (r *RecordingRunner) Run(stdOut io.Writer, stdErr io.Writer, commandPath string, arg ...string) int {
	r.Cmd = commandPath
	r.Args = arg
	return 0
}

func (r *RecordingRunner) Output(command string, args ...string) ([]byte, error) {
	r.Cmd = command
	r.Args = args
	return ocClusterUpHelp, nil
}

var testDir string
var testRunner *RecordingRunner
var testMachineConfig = cluster.MachineConfig{
	OpenShiftVersion: "v1.3.1",
}
var testIp = "192.168.99.42"

func TestStartClusterUpNoFlags(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer SetRunner(util.RealRunner{})
	defer viper.Reset()

	clusterUp(&testMachineConfig, testIp)

	expectedOc := filepath.Join(testDir, "cache", "oc", testMachineConfig.OpenShiftVersion, constants.OC_BINARY_NAME)
	if testRunner.Cmd != expectedOc {
		t.Errorf("Expected command '%s'. Received '%s'", expectedOc, testRunner.Cmd)
	}

	expectedArguments := []string{"cluster", "up", "--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", hostVolumesDirectory,
		"--routing-suffix", testIp + ".nip.io",
	}
	for i, v := range testRunner.Args {
		if v != expectedArguments[i] {
			t.Errorf("Expected argument '%s'. Received '%s'", expectedArguments[i], v)
		}
	}
}

func TestStartClusterUpWithOverrideHostConfigDirFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer SetRunner(util.RealRunner{})
	defer viper.Reset()

	viper.Set("host-config-dir", "/var/tmp/foo")
	clusterUp(&testMachineConfig, testIp)

	expectedArguments := []string{"cluster", "up", "--use-existing-config",
		"--host-config-dir", "/var/tmp/foo",
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", hostVolumesDirectory,
		"--routing-suffix", testIp + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)
}

func TestStartClusterUpWithOverrideHostDataDirFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer SetRunner(util.RealRunner{})
	defer viper.Reset()

	viper.Set("host-data-dir", "/var/tmp/foo")
	clusterUp(&testMachineConfig, testIp)

	expectedArguments := []string{"cluster", "up", "--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", "/var/tmp/foo",
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", hostVolumesDirectory,
		"--routing-suffix", testIp + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)
}

func TestStartClusterUpWithOverrideHostVolumesDirFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer SetRunner(util.RealRunner{})
	defer viper.Reset()

	viper.Set("host-volumes-dir", "/var/tmp/foo")
	clusterUp(&testMachineConfig, testIp)

	expectedArguments := []string{"cluster", "up", "--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", "/var/tmp/foo",
		"--routing-suffix", testIp + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)
}

func TestStartClusterUpWithOverrideHostPvDirFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer SetRunner(util.RealRunner{})
	defer viper.Reset()

	viper.Set("host-pv-dir", "/var/tmp/foo")
	clusterUp(&testMachineConfig, testIp)

	expectedArguments := []string{"cluster", "up", "--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", "/var/tmp/foo",
		"--host-volumes-dir", hostVolumesDirectory,
		"--routing-suffix", testIp + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)
}

func TestStartClusterUpWithFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer SetRunner(util.RealRunner{})
	defer viper.Reset()

	viper.Set("public-hostname", "foobar")
	viper.Set("skip-registry-check", "true")
	clusterUp(&testMachineConfig, testIp)

	expectedArguments := []string{"cluster", "up", "--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", hostVolumesDirectory,
		"--public-hostname", "foobar",
		"--routing-suffix", testIp + ".nip.io",
		"--skip-registry-check", "true"}
	assertCommandLineArguments(expectedArguments, t)
}

func TestShellProxyVariableWithProxyFlag(t *testing.T) {
	defer viper.Reset()

	expectedShellProxyVariable := "http_proxy=http://localhost:3128 https_proxy=https://localhost:3128 no_proxy=localhost,127.0.0.1"

	viper.Set("http-proxy", "http://localhost:3128")
	viper.Set("https-proxy", "https://localhost:3128")
	viper.Set("no-proxy", "localhost,127.0.0.1")

	setShellProxy()

	if shellProxyEnv != expectedShellProxyVariable {
		t.Fatalf("Shell variable doesn't set properly expected %s got %s", expectedShellProxyVariable, shellProxyEnv)
	}
}

func TestDockerEnvWithProxyFlag(t *testing.T) {
	defer viper.Reset()

	expectedDockerEnv := []string{"HTTP_PROXY=http://localhost:3128",
		fmt.Sprintf("NO_PROXY=%s", updateNoProxyForDocker())}

	viper.Set("http-proxy", "http://localhost:3128")

	setDockerProxy()

	assertCompareSlice(expectedDockerEnv, dockerEnv, t)
}

func TestClusterUpWithProxyFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer SetRunner(util.RealRunner{})
	defer viper.Reset()

	viper.Set("http-proxy", "http://localhost:3128")
	viper.Set("https-proxy", "https://localhost:3128")
	viper.Set("no-proxy", "10.0.0.1")

	setOcProxy()
	// To make sure oc download doesn't take localhost:3128 proxy in account
	proxyUrl = ""

	clusterUp(&testMachineConfig, testIp)

	expectedArguments := []string{"cluster", "up", "--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", hostVolumesDirectory,
		"--http-proxy", "http://localhost:3128",
		"--https-proxy", "https://localhost:3128",
		"--no-proxy", "10.0.0.1",
		"--routing-suffix", testIp + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)

}

func TestStartClusterUpWithOpenShiftEnv(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer SetRunner(util.RealRunner{})
	defer viper.Reset()

	viper.Set("openshift-env", "HTTP_PROXY=http://localhost:3128,HTTP_PROXY_USER=foo,HTTP_PROXY_PASS=bar")
	clusterUp(&testMachineConfig, testIp)

	expectedArguments := []string{"cluster", "up", "--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", hostVolumesDirectory,
		"--env", "HTTP_PROXY=http://localhost:3128,HTTP_PROXY_USER=foo,HTTP_PROXY_PASS=bar",
		"--routing-suffix", testIp + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)
}

func TestNoExplicitRouteSuffixDefaultsToNip(t *testing.T) {
	defer viper.Reset()
	setDefaultRoutingPrefix(testIp)

	expectedRoutingSuffix := testIp + ".nip.io"
	if viper.Get(routingSuffix) != expectedRoutingSuffix {
		t.Fatalf("Expected argument '%s'. Received '%s'", expectedRoutingSuffix, viper.Get(routingSuffix))
	}
}

func TestExplicitRouteSuffixGetApplied(t *testing.T) {
	explicitRoutingSuffix := "acme.com"

	viper.Set(routingSuffix, explicitRoutingSuffix)
	defer viper.Reset()

	setDefaultRoutingPrefix(testIp)

	if viper.Get(routingSuffix) != explicitRoutingSuffix {
		t.Fatalf("Expected argument '%s'. Received '%s'", explicitRoutingSuffix, viper.Get(routingSuffix))
	}
}

func assertCompareSlice(expectedArguments []string, recievedArguments []string, t *testing.T) {
	if len(expectedArguments) > len(recievedArguments) {
		t.Errorf("Expected more arguments than received. Expected: '%s'. Got: '%s'", expectedArguments, recievedArguments)
	}

	if len(expectedArguments) < len(recievedArguments) {
		t.Errorf("Received more arguments than expected. Expected: '%s'. Got '%s'", expectedArguments, recievedArguments)
	}

	for i, v := range recievedArguments {
		if v != expectedArguments[i] {
			t.Errorf("Expected argument '%s'. Received '%s'", expectedArguments[i], v)
		}
	}
}

func assertCommandLineArguments(expectedArguments []string, t *testing.T) {
	if len(expectedArguments) > len(testRunner.Args) {
		t.Errorf("Expected more arguments than received. Expected: '%s'. Got: '%s'", expectedArguments, testRunner.Args)
	}

	if len(expectedArguments) < len(testRunner.Args) {
		t.Errorf("Received more arguments than expected. Expected: '%s'. Got '%s'", expectedArguments, testRunner.Args)
	}

	for i, v := range testRunner.Args {
		if v != expectedArguments[i] {
			t.Errorf("Expected argument '%s'. Received '%s'", expectedArguments[i], v)
		}
	}
}

func setUp(t *testing.T) {
	var err error
	testDir, err = ioutil.TempDir("", "minishift-test-start-cmd-")
	if err != nil {
		t.Error(err)
	}

	machinesDirPath := filepath.Join(testDir, "machines")
	os.Mkdir(machinesDirPath, 0755)
	instanceState.InstanceConfig, err = instanceState.NewInstanceConfig(filepath.Join(machinesDirPath, "fake-machines.json"))
	if err != nil {
		t.Error(err)
	}

	constants.Minipath = testDir

	client := http.DefaultClient
	client.Transport = minitesting.NewMockRoundTripper()

	testRunner = &RecordingRunner{}
	SetRunner(testRunner)

	// Set default value for host config and data
	viper.Set("host-config-dir", hostConfigDirectory)
	viper.Set("host-data-dir", hostDataDirectory)
	viper.Set("host-volumes-dir", hostVolumesDirectory)
	viper.Set("host-pv-dir", hostPvDirectory)

	provision.SetDetector(&tests.MockDetector{&tests.MockProvisioner{Provisioned: true}})
}
