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

package oc

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"runtime"
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

	expectedOptions = []string{"base-dir", "enable",
		"forward-ports", "http-proxy", "https-proxy",
		"image", "no-proxy", "public-hostname", "routing-suffix", "server-loglevel",
		"skip-registry-check", "write-config"}
)

func Test_invalid_oc_path_returns_error(t *testing.T) {
	invalidPath := "/snafu"
	_, err := NewOcRunner(invalidPath, "")
	assert.Error(t, err, "An error should have been returned for creating on OcRunner against an invalid path")

	expectedError := fmt.Sprintf(invalidOcPathError, invalidPath)
	assert.EqualError(t, err, expectedError)
}

func Test_invalid_kube_path_returns_error(t *testing.T) {
	// for now it is enough to just pass a file, there are no checks for name of the file or whether
	// it is executable
	tmpOc, err := ioutil.TempFile("", "oc")
	assert.NoError(t, err)

	defer os.Remove(tmpOc.Name()) // clean up

	invalidPath := "/snafu"
	_, err = NewOcRunner(tmpOc.Name(), invalidPath)
	assert.Error(t, err, "An error should have been returned for creating on OcRunner against an invalid path")

	expectedError := fmt.Sprintf(invalidKubeConfigPathError, invalidPath)
	assert.EqualError(t, err, expectedError)
}

func Test_quotes_in_commands_get_properly_parsed(t *testing.T) {
	fakeRunner := FakeRunner{}
	expectedCommand := "/foo/bar"
	ocRunner := OcRunner{
		OcPath:         expectedCommand,
		KubeConfigPath: "/kube/config",
		Runner:         &fakeRunner,
	}

	ocRunner.Run("adm new-project foo --description='Bar Baz'", nil, nil)

	assert.Equal(t, expectedCommand, fakeRunner.cmd)

	expectedLength := 5
	assert.Len(t, fakeRunner.args, expectedLength)

	expectedArgs := []string{"--config=/kube/config", "adm", "new-project", "foo", "--description='Bar Baz'"}
	for i, expectedArg := range expectedArgs {
		assert.Equal(t, expectedArg, fakeRunner.args[i])
	}
}

func TestParseOcHelpCommand(t *testing.T) {
	recievedOptions := parseOcHelpCommand(ocClusterUpHelp)
	assert.EqualValues(t, expectedOptions, recievedOptions)
}

func TestFlagExist(t *testing.T) {
	assert.True(t, flagExist(expectedOptions, "image"))
	assert.False(t, flagExist(expectedOptions, "proxy"))
}

func Test_get_kubeconfig_global_path(t *testing.T) {
	type testdata struct {
		KubeConfig string
		Expected   string
	}
	var tests []testdata

	if runtime.GOOS == "windows" {
		tests = []testdata{{KubeConfig: "/tmp/john/config", Expected: "/tmp/john/config"},
			{KubeConfig: "/tmp/john/config;/tmp/foo/config", Expected: "/tmp/foo/config"},
			{KubeConfig: "/tmp/john/config;/tmp/foo/config;/tmp/bar/config", Expected: "/tmp/bar/config"},
			{KubeConfig: ";;", Expected: getUserKubeConfigLocation()},
			{KubeConfig: ";/tmp/foo/config;", Expected: "/tmp/foo/config"},
			{KubeConfig: ";/tmp/foo/config:s", Expected: "/tmp/foo/config:s"},
			{KubeConfig: ";/tmp/foo/config", Expected: "/tmp/foo/config"},
			{KubeConfig: ":/tmp/foo/config", Expected: ":/tmp/foo/config"},
			{KubeConfig: "", Expected: getUserKubeConfigLocation()},
		}
	} else {
		tests = []testdata{{KubeConfig: "/tmp/john/config", Expected: "/tmp/john/config"},
			{KubeConfig: "/tmp/john/config:/tmp/foo/config", Expected: "/tmp/foo/config"},
			{KubeConfig: "/tmp/john/config;/tmp/foo/config:/tmp/bar/config", Expected: "/tmp/bar/config"},
			{KubeConfig: "::", Expected: getUserKubeConfigLocation()},
			{KubeConfig: ":/tmp/foo/config:", Expected: "/tmp/foo/config"},
			{KubeConfig: ":/tmp/foo/config;s", Expected: "/tmp/foo/config;s"},
			{KubeConfig: ":/tmp/foo/config:", Expected: "/tmp/foo/config"},
			{KubeConfig: ";/tmp/foo/config", Expected: ";/tmp/foo/config"},
			{KubeConfig: "", Expected: getUserKubeConfigLocation()},
		}
	}
	for _, data := range tests {
		os.Setenv("KUBECONFIG", data.KubeConfig)
		expected, err := getGlobalKubeConfigPath()
		assert.NoError(t, err)
		assert.Equal(t, expected, data.Expected)
	}
}

type FakeRunner struct {
	cmd  string
	args []string
}

func (r *FakeRunner) Output(command string, args ...string) ([]byte, error) {
	return nil, nil
}

func (r *FakeRunner) Run(stdOut io.Writer, stdErr io.Writer, commandPath string, args ...string) int {
	r.cmd = commandPath
	r.args = args
	return 0
}

func getUserKubeConfigLocation() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, ".kube", "config")
}
