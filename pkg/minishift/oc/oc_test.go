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
	"testing"

	"github.com/stretchr/testify/assert"
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
      --create-machine=false: Create a Docker machine if one doesn't exist
      --docker-machine='': Specify the Docker machine to use
  -e, --env=[]: Specify a key-value pair for an environment variable to set on OpenShift container
      --forward-ports=true: Use Docker port-forwarding to communicate with origin container. Requires 'socat' locally.
      --host-config-dir='/var/lib/origin/openshift.local.config': Directory on Docker host for OpenShift configuration
      --host-data-dir='': Directory on Docker host for OpenShift data. If not specified, etcd data will not be persisted
on the host.
      --host-volumes-dir='/var/lib/origin/openshift.local.volumes': Directory on Docker host for OpenShift volumes
      --image='openshift/origin': Specify the images to use for OpenShift
      --logging=false: Install logging (experimental)
      --metrics=false: Install metrics (experimental)
      --public-hostname='': Public hostname for OpenShift cluster
      --routing-suffix='': Default suffix for server routes
      --server-loglevel=0: Log level for OpenShift server
      --skip-registry-check=false: Skip Docker daemon registry check
      --use-existing-config=false: Use existing configuration if present
      --version='': Specify the tag for OpenShift images

Use "oc options" for a list of global command-line options (applies to all commands).`)

	expectedOptions = []string{"create-machine", "docker-machine", "env",
		"forward-ports", "host-config-dir", "host-data-dir", "host-volumes-dir",
		"image", "logging", "metrics", "public-hostname", "routing-suffix", "server-loglevel",
		"skip-registry-check", "use-existing-config", "version"}
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
