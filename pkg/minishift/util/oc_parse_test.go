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

package util

import (
	"testing"
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

func TestParseOcHelpCommand(t *testing.T) {
	recievedOptions := ParseOcHelpCommand(ocClusterUpHelp)
	assertCompareSlice(expectedOptions, recievedOptions, t)
}

func TestFlagExist(t *testing.T) {
	if !FlagExist(expectedOptions, "image") {
		t.Fatal("image flag exist but returned false")
	}

	if FlagExist(expectedOptions, "proxy") {
		t.Fatal("proxy flag doesn't exist but returned true")
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
