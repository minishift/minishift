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

	"github.com/minishift/minishift/pkg/util"
	"os"
	"io/ioutil"
	"github.com/minishift/minishift/pkg/minikube/tests"
	"github.com/docker/machine/libmachine/provision"
	"github.com/spf13/viper"
	"path/filepath"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"net/http"
	"github.com/minishift/minishift/pkg/minikube/constants"
)

type RecordingRunner struct {
	Cmd  string
	Args []string
}

func (r *RecordingRunner) Run(command string, args ...string) error {
	r.Cmd = command
	r.Args = args
	return nil
}

var testDir string
var testRunner *RecordingRunner
var testMachineConfig = cluster.MachineConfig {
	OpenShiftVersion: "v1.3.1",
}

func TestStartClusterUpNoFlags(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer SetRunner(util.RealRunner{})
	defer viper.Reset()

	clusterUp(&testMachineConfig)

	expectedOc := filepath.Join(testDir, "cache", "oc", testMachineConfig.OpenShiftVersion, constants.OC_BINARY_NAME)
	if testRunner.Cmd != expectedOc {
		t.Errorf("Expected command '%s'. Got '%s'", expectedOc, testRunner.Cmd)
	}

	expectedArguments := []string{"cluster", "up"}
	for i, v := range testRunner.Args {
		if v != expectedArguments[i] {
			t.Errorf("Expected argument '%s'. Got '%s'", expectedArguments[i], v)
		}
	}
}

func TestStartClusterUpWithFlag(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer SetRunner(util.RealRunner{})
	defer viper.Reset()

	viper.Set("public-hostname", "foobar")
	viper.Set("skip-registry-check", "true")
	clusterUp(&testMachineConfig)

	expectedArguments := []string{"cluster", "up", "--public-hostname", "foobar", "--skip-registry-check", "true"}
	assertCommandLineArguments(expectedArguments, t)
}

func TestStartClusterUpWithOpenShiftEnv(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)
	defer minitesting.ResetDefaultRoundTripper()
	defer SetRunner(util.RealRunner{})
	defer viper.Reset()

	viper.Set("openshift-env", "HTTP_PROXY=http://localhost:3128,HTTP_PROXY_USER=foo,HTTP_PROXY_PASS=bar")
	clusterUp(&testMachineConfig)

	expectedArguments := []string{"cluster", "up", "--env", "HTTP_PROXY=http://localhost:3128,HTTP_PROXY_USER=foo,HTTP_PROXY_PASS=bar"}
	assertCommandLineArguments(expectedArguments, t)
}

func assertCommandLineArguments(expectedArguments []string, t *testing.T) {
	if len(expectedArguments) > len(testRunner.Args) {
		t.Errorf("Expected more arguments than were received. Expected: '%s'. Got: '%s'", expectedArguments, testRunner.Args)
	}

	if len(expectedArguments) < len(testRunner.Args) {
		t.Errorf("Got more arguments than were expected. Expected: '%s'. Got '%s'", expectedArguments, testRunner.Args)
	}

	for i, v := range testRunner.Args {
		if v != expectedArguments[i] {
			t.Errorf("Expected argument '%s'. Got '%s'", expectedArguments[i], v)
		}
	}
}

func setUp(t *testing.T) {
	var err error
	testDir, err = ioutil.TempDir("", "minishift-test-start-cmd-")
	if err != nil {
		t.Error(err)
	}
	SetMinishiftDir(testDir)

	//os.MkdirAll(filepath.Join(testDir, "certs"), 0777)
	//
	//isoCacheDir := filepath.Join(testDir, "cache", "iso")
	//os.MkdirAll(isoCacheDir, 0777)
	//os.OpenFile(filepath.Join(isoCacheDir, "boot2docker.iso"), os.O_RDONLY | os.O_CREATE, 0666)
	//
	//machinesDir := filepath.Join(testDir, "machines", "minishift")
	//os.MkdirAll(machinesDir, 0777)

	client := http.DefaultClient
	client.Transport = minitesting.NewMockRoundTripper()

	testRunner = &RecordingRunner{}
	SetRunner(testRunner)

	provision.SetDetector(&tests.MockDetector{&tests.MockProvisioner{Provisioned: true}})
}
