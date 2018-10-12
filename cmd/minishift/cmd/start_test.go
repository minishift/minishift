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

	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/machine/libmachine/provision"
	"github.com/stretchr/testify/assert"

	"fmt"
	"sort"
	"strings"

	"bytes"

	"github.com/minishift/minishift/cmd/testing/cli"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minikube/tests"
	"github.com/minishift/minishift/pkg/minishift/clusterup"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	pkgTest "github.com/minishift/minishift/pkg/testing"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/viper"
)

var (
	ocClusterUpHelp = []byte(`Starts an OpenShift cluster using Docker containers, provisioning a registry, router, initial templates, and a default
project. 

This command will attempt to use an existing connection to a Docker daemon. Before running the command, ensure that you
can execute docker commands successfully (i.e. 'docker ps'). 

By default, the OpenShift cluster will be setup to use a routing suffix that ends in nip.io. This is to allow dynamic
host names to be created for routes. An alternate routing suffix can be specified using the --routing-suffix flag. 

A public hostname can also be specified for the server with the --public-hostname flag.

Usage:
  oc cluster up [flags]

Examples:
  # Start OpenShift using a specific public host name
  oc cluster up --public-hostname=my.address.example.com

Options:
      --base-dir='': Directory on Docker host for cluster up configuration
      --enable=[*]: A list of components to enable.  '*' enables all on-by-default components, 'foo' enables the
component named 'foo', '-foo' disables the component named 'foo'.
All components: automation-service-broker, centos-imagestreams, persistent-volumes, registry, rhel-imagestreams, router,
sample-templates, service-catalog, template-service-broker, web-console
Disabled-by-default components: automation-service-broker, rhel-imagestreams, service-catalog, template-service-broker
      --forward-ports=false: Use Docker port-forwarding to communicate with origin container. Requires 'socat' locally.
      --http-proxy='': HTTP proxy to use for master and builds
      --https-proxy='': HTTPS proxy to use for master and builds
      --image='openshift/origin-${component}:${version}': Specify the images to use for OpenShift
      --no-proxy=[]: List of hosts or subnets for which a proxy should not be used
      --public-hostname='': Public hostname for OpenShift cluster
      --routing-suffix='': Default suffix for server routes
      --server-loglevel=0: Log level for OpenShift server
      --skip-registry-check=false: Skip Docker daemon registry check
      --write-config=false: Write the configuration files into host config dir

Use "oc options" for a list of global command-line options (applies to all commands).`)
	dockerSubnetForTest = "172.17.0.0/16"
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
		OpenShiftVersion: "v3.11.0",
		Ip:               "192.168.99.42",
		OcPath:           "/home/john/.minishift/cache/oc/v3.11.0/linux/oc",
		RoutingSuffix:    "192.168.99.42.nip.io",
	}
	testDir    string
	testRunner *RecordingRunner
	tee        *pkgTest.Tee
)

func TestStartClusterUpNoFlags(t *testing.T) {
	setUp(t)
	defer tearDown()

	clusterUpParams := determineClusterUpParameters(testConfig, dockerSubnetForTest)
	clusterup.ClusterUp(testConfig, clusterUpParams)

	assert.Equal(t, testConfig.OcPath, testRunner.Cmd)

	expectedArguments := []string{
		"cluster",
		"up",
		"--use-existing-config",
		"--routing-suffix", testConfig.Ip + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)
}

func TestStartClusterUpWithFlag(t *testing.T) {
	setUp(t)
	defer tearDown()

	viper.Set("public-hostname", "foobar")
	viper.Set("skip-registry-check", "true")

	clusterUpParams := determineClusterUpParameters(testConfig, dockerSubnetForTest)
	clusterup.ClusterUp(testConfig, clusterUpParams)

	expectedArguments := []string{
		"cluster",
		"up",
		"--use-existing-config",
		"--public-hostname", "foobar",
		"--routing-suffix", testConfig.Ip + ".nip.io",
		"--skip-registry-check", "true"}
	assertCommandLineArguments(expectedArguments, t)
}

func TestClusterUpWithProxyFlag(t *testing.T) {
	setUp(t)
	defer tearDown()

	viper.Set("http-proxy", "http://localhost:3128")
	viper.Set("https-proxy", "https://localhost:3128")
	viper.Set("no-proxy", "10.0.0.1")

	clusterUpParams := determineClusterUpParameters(testConfig, dockerSubnetForTest)
	clusterup.ClusterUp(testConfig, clusterUpParams)

	expectedArguments := []string{
		"cluster",
		"up",
		"--use-existing-config",
		"--http-proxy", "http://localhost:3128",
		"--https-proxy", "https://localhost:3128",
		"--no-proxy", "10.0.0.1",
		"--routing-suffix", testConfig.Ip + ".nip.io",
	}
	assertCommandLineArguments(expectedArguments, t)

}

func TestCheckMemorySize(t *testing.T) {
	var sizeTests = []struct {
		in  string
		out int
	}{
		{"4096", 4096},
		{"4096M", 4096},
		{"4G", 4096},
		{"4GB", 4096},
	}

	for _, sizeTest := range sizeTests {
		size := calculateMemorySize(sizeTest.in)

		assert.Equal(t, sizeTest.out, size)
	}
}

func TestCheckDiskSize(t *testing.T) {
	var sizeTests = []struct {
		in  string
		out int
	}{
		{"20000", 20000},
		{"20000M", 20000},
		{"20G", 20000},
		{"20GB", 20000},
	}

	for _, sizeTest := range sizeTests {
		size := calculateDiskSize(sizeTest.in)

		assert.Equal(t, sizeTest.out, size)
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
		{"https://my.custom.url/Myiso.iso", "https://my.custom.url/Myiso.iso"},
		{"file://somewhere/on/disk", "file://somewhere/on/disk"},
		{"file://somewhere/On/disk", "file://somewhere/On/disk"},
	}

	for _, isoTest := range isoTests {
		isoUrl := determineIsoUrl(isoTest.in)
		assert.Equal(t, isoTest.out, isoUrl)
	}
}

func TestDetermineIsoUrlWithInvalidName(t *testing.T) {
	tee = cli.CreateTee(t, true)
	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 1, unsupportedIsoUrlFormat))
	defer tearDown()

	determineIsoUrl("foo")
}

func Test_getslice_withConfig(t *testing.T) {
	viper.SetConfigType("json")
	defer viper.Reset()
	var confFile = []byte(`
{
    "bar": [
        "hello",
        "world"
    ]
}
	`)

	viper.ReadConfig(bytes.NewBuffer(confFile))

	expectedSlice := []string{"hello", "world"}
	assert.Len(t, expectedSlice, len(getSlice("bar")))

	for i, v := range getSlice("bar") {
		assert.Equal(t, expectedSlice[i], v)
	}
}

func Test_getslice_withCommandLine(t *testing.T) {
	viper.Set("foo", []string{"hello", "world"})
	defer viper.Reset()

	expectedSlice := []string{"hello", "world"}
	assert.Len(t, expectedSlice, len(getSlice("foo")))

	for i, v := range getSlice("foo") {
		assert.Equal(t, expectedSlice[i], v)
	}
}

func assertCommandLineArguments(expectedArguments []string, t *testing.T) {
	assert.Len(t, expectedArguments, len(testRunner.Args))

	sort.Strings(testRunner.Args)
	sort.Strings(expectedArguments)

	for i, v := range testRunner.Args {
		assert.Equal(t, expectedArguments[i], v)
	}
}

func setUp(t *testing.T) {
	var err error
	testDir, err = ioutil.TempDir("", "minishift-test-start-cmd-")
	assert.NoError(t, err, "Error creating temporary directory")

	machinesDirPath := filepath.Join(testDir, "machines")
	os.Mkdir(machinesDirPath, 0755)
	instanceState.InstanceStateConfig, err = instanceState.NewInstanceStateConfig(filepath.Join(machinesDirPath, "fake-machines.json"))

	assert.NoError(t, err, "Error getting new instance config")
	constants.Minipath = testDir

	testRunner = &RecordingRunner{}

	provision.SetDetector(&tests.MockDetector{&tests.MockProvisioner{Provisioned: true}})

	atexit.RegisterExitHandler(cli.PreventExitWithNonZeroExitCode(t))
}

func tearDown() {
	if tee != nil {
		tee.Close()
		tee = nil
	}

	if testDir != "" {
		os.RemoveAll(testDir)
		testDir = ""
	}

	viper.Reset()
	atexit.ClearExitHandler()

	if r := recover(); r != nil {
		reason := fmt.Sprint(r)
		if reason != atexit.ExitHandlerPanicMessage {
			fmt.Println("Recovered from panic:", r)
		}
	}
}
