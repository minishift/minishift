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

	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/machine/libmachine/provision"

	"github.com/minishift/minishift/cmd/minishift/cmd/config"

	"fmt"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minikube/tests"
	"github.com/minishift/minishift/pkg/minishift/clusterup"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/testing/cli"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/viper"
	"sort"
	"strings"
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

  # Specify which set of image streams to use
  oc cluster up --image-streams=centos7

Options:
      --create-machine=false: Create a Docker machine if one doesn't exist
      --docker-machine='': Specify the Docker machine to use
  -e, --env=[]: Specify a key-value pair for an environment variable to set on OpenShift container
      --forward-ports=true: Use Docker port-forwarding to communicate with origin container. Requires 'socat' locally.
      --host-config-dir='/var/lib/origin/openshift.local.config': Directory on Docker host for OpenShift configuration
      --host-data-dir='': Directory on Docker host for OpenShift data. If not specified, etcd data will not be persisted
on the host.
      --host-pv-dir='/var/lib/origin/openshift.local.pv': Directory on host for OpenShift persistent volumes
      --host-volumes-dir='/var/lib/origin/openshift.local.volumes': Directory on Docker host for OpenShift volumes
      --http-proxy='': HTTP proxy to use for master and builds
      --https-proxy='': HTTPS proxy to use for master and builds
      --image='openshift/origin': Specify the images to use for OpenShift
      --image-streams='centos7': Specify which image streams to use, centos7|rhel7
      --logging=false: Install logging (experimental)
      --metrics=false: Install metrics (experimental)
      --no-proxy=[]: List of hosts or subnets for which a proxy should not be used
      --public-hostname='': Public hostname for OpenShift cluster
      --routing-suffix='': Default suffix for server routes
      --server-loglevel=0: Log level for OpenShift server
      --skip-registry-check=false: Skip Docker daemon registry check
      --use-existing-config=false: Use existing configuration if present
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

	if strings.Join(arg, " ") == "cluster up -h" {
		stdOut.Write(ocClusterUpHelp)
	}

	return 0
}

func (r *RecordingRunner) Output(command string, args ...string) ([]byte, error) {
	r.Cmd = command
	r.Args = args
	return ocClusterUpHelp, nil
}

var (
	testConfig = &clusterup.ClusterUpConfig{
		OpenShiftVersion: "v1.5.1",
		Ip:               "192.168.99.42",
		OcPath:           "/home/john/.minishift/cache/oc/1.5.1/oc",
	}
	testDir    string
	testRunner *RecordingRunner
)

func TestStartClusterUpNoFlags(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer viper.Reset()

	clusterUpParams := determineClusterUpParameters(testConfig)
	clusterup.ClusterUp(testConfig, clusterUpParams, testRunner)

	if testRunner.Cmd != testConfig.OcPath {
		t.Errorf("Expected command '%s'. Received '%s'", testConfig.OcPath, testRunner.Cmd)
	}

	expectedArguments := []string{
		"cluster",
		"up",
		"--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", hostVolumesDirectory,
		"--routing-suffix", testConfig.Ip + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)
}

func TestStartClusterUpWithOverrideHostConfigDirFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer viper.Reset()

	viper.Set("host-config-dir", "/var/tmp/foo")

	clusterUpParams := determineClusterUpParameters(testConfig)
	clusterup.ClusterUp(testConfig, clusterUpParams, testRunner)

	expectedArguments := []string{
		"cluster",
		"up",
		"--use-existing-config",
		"--host-config-dir", "/var/tmp/foo",
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", hostVolumesDirectory,
		"--routing-suffix", testConfig.Ip + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)
}

func TestStartClusterUpWithOverrideHostDataDirFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer viper.Reset()

	viper.Set("host-data-dir", "/var/tmp/foo")

	clusterUpParams := determineClusterUpParameters(testConfig)
	clusterup.ClusterUp(testConfig, clusterUpParams, testRunner)

	expectedArguments := []string{
		"cluster",
		"up",
		"--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", "/var/tmp/foo",
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", hostVolumesDirectory,
		"--routing-suffix", testConfig.Ip + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)
}

func TestStartClusterUpWithOverrideHostVolumesDirFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer viper.Reset()

	viper.Set("host-volumes-dir", "/var/tmp/foo")

	clusterUpParams := determineClusterUpParameters(testConfig)
	clusterup.ClusterUp(testConfig, clusterUpParams, testRunner)

	expectedArguments := []string{
		"cluster",
		"up",
		"--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", "/var/tmp/foo",
		"--routing-suffix", testConfig.Ip + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)
}

func TestStartClusterUpWithOverrideHostPvDirFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer viper.Reset()

	viper.Set("host-pv-dir", "/var/tmp/foo")

	clusterUpParams := determineClusterUpParameters(testConfig)
	clusterup.ClusterUp(testConfig, clusterUpParams, testRunner)

	expectedArguments := []string{
		"cluster",
		"up",
		"--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", "/var/tmp/foo",
		"--host-volumes-dir", hostVolumesDirectory,
		"--routing-suffix", testConfig.Ip + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)
}

func TestStartClusterUpWithFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer viper.Reset()

	viper.Set("public-hostname", "foobar")
	viper.Set("skip-registry-check", "true")

	clusterUpParams := determineClusterUpParameters(testConfig)
	clusterup.ClusterUp(testConfig, clusterUpParams, testRunner)

	expectedArguments := []string{
		"cluster",
		"up",
		"--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", hostVolumesDirectory,
		"--public-hostname", "foobar",
		"--routing-suffix", testConfig.Ip + ".nip.io",
		"--skip-registry-check", "true"}
	assertCommandLineArguments(expectedArguments, t)
}

func TestClusterUpWithProxyFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer viper.Reset()

	viper.Set("http-proxy", "http://localhost:3128")
	viper.Set("https-proxy", "https://localhost:3128")
	viper.Set("no-proxy", "10.0.0.1")

	clusterUpParams := determineClusterUpParameters(testConfig)
	clusterup.ClusterUp(testConfig, clusterUpParams, testRunner)

	expectedArguments := []string{
		"cluster",
		"up",
		"--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", hostVolumesDirectory,
		"--http-proxy", "http://localhost:3128",
		"--https-proxy", "https://localhost:3128",
		"--no-proxy", "10.0.0.1",
		"--routing-suffix", testConfig.Ip + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)

}

func TestStartClusterUpWithOpenShiftEnv(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer viper.Reset()

	viper.Set("openshift-env", "HTTP_PROXY=http://localhost:3128,HTTP_PROXY_USER=foo,HTTP_PROXY_PASS=bar")

	clusterUpParams := determineClusterUpParameters(testConfig)
	clusterup.ClusterUp(testConfig, clusterUpParams, testRunner)

	expectedArguments := []string{
		"cluster",
		"up",
		"--use-existing-config",
		"--host-config-dir", hostConfigDirectory,
		"--host-data-dir", hostDataDirectory,
		"--host-pv-dir", hostPvDirectory,
		"--host-volumes-dir", hostVolumesDirectory,
		"--env", "HTTP_PROXY=http://localhost:3128,HTTP_PROXY_USER=foo,HTTP_PROXY_PASS=bar",
		"--routing-suffix", testConfig.Ip + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)
}

func TestNoExplicitRouteSuffixDefaultsToNip(t *testing.T) {
	defer viper.Reset()

	setDefaultRoutingPrefix(testConfig.Ip)
	expectedRoutingSuffix := testConfig.Ip + ".nip.io"

	if viper.Get(config.RoutingSuffix.Name) != expectedRoutingSuffix {
		t.Fatalf("Expected argument '%s'. Received '%s'", expectedRoutingSuffix, viper.Get(config.RoutingSuffix.Name))
	}
}

func TestExplicitRouteSuffixGetApplied(t *testing.T) {
	explicitRoutingSuffix := "acme.com"

	viper.Set(config.RoutingSuffix.Name, explicitRoutingSuffix)
	defer viper.Reset()

	setDefaultRoutingPrefix(testConfig.Ip)

	if viper.Get(config.RoutingSuffix.Name) != explicitRoutingSuffix {
		t.Fatalf("Expected argument '%s'. Received '%s'", explicitRoutingSuffix, viper.Get(config.RoutingSuffix.Name))
	}
}

func TestDetermineIsoUrl(t *testing.T) {
	var isoTests = []struct {
		in  string
		out string
	}{
		{"", constants.DefaultB2dIsoUrl},
		{"b2d", constants.DefaultB2dIsoUrl},
		{"B2D", constants.DefaultB2dIsoUrl},
		{"centos", constants.DefaultCentOsIsoUrl},
		{"CentOs", constants.DefaultCentOsIsoUrl},
		{"http://my.custom.url/myiso.iso", "http://my.custom.url/myiso.iso"},
		{"https://my.custom.url/myiso.iso", "https://my.custom.url/myiso.iso"},
		{"file://somewhere/on/disk", "file://somewhere/on/disk"},
	}

	for _, isoTest := range isoTests {
		isoUrl := determineIsoUrl(isoTest.in)
		if isoUrl != isoTest.out {
			t.Errorf("Expected '%s' as ISO URL for input '%s'. Got '%s'.", isoTest.out, isoTest.in, isoUrl)
		}
	}
}

func TestDetermineIsoUrlWithInvalidName(t *testing.T) {
	tee := cli.CreateTee(t, true)
	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 1, unsupportedIsoUrlFormat))
	defer tearDown(tee)

	determineIsoUrl("foo")
}

func assertCommandLineArguments(expectedArguments []string, t *testing.T) {
	if len(expectedArguments) > len(testRunner.Args) {
		t.Errorf("Expected more arguments than received. Expected: '%s'. Got: '%s'", expectedArguments, testRunner.Args)
	}

	if len(expectedArguments) < len(testRunner.Args) {
		t.Errorf("Received more arguments than expected. Expected: '%s'. Got '%s'", expectedArguments, testRunner.Args)
	}

	sort.Strings(testRunner.Args)
	sort.Strings(expectedArguments)

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

	testRunner = &RecordingRunner{}

	// Set default value for host config and data
	viper.Set("host-config-dir", hostConfigDirectory)
	viper.Set("host-data-dir", hostDataDirectory)
	viper.Set("host-volumes-dir", hostVolumesDirectory)
	viper.Set("host-pv-dir", hostPvDirectory)

	provision.SetDetector(&tests.MockDetector{&tests.MockProvisioner{Provisioned: true}})

	atexit.RegisterExitHandler(cli.PreventExitWithNonZeroExitCode(t))
}

func tearDown(tee *cli.Tee) {
	tee.Close()
	viper.Reset()
	atexit.ClearExitHandler()
	if r := recover(); r != nil {
		reason := fmt.Sprint(r)
		if reason != atexit.ExitHandlerPanicMessage {
			fmt.Println("Recovered from panic:", r)
		}
	}
}
