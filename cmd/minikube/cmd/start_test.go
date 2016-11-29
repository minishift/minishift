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

	"net/http"
	"github.com/minishift/minishift/pkg/util"
	"os"
	"path/filepath"
	"io/ioutil"
)

type RecordingRunner struct{
	Cmd  string
	Args []string
}

func (r *RecordingRunner) Run(command string, args ...string) error {
	r.Cmd = command
	r.Args = args
	return nil
}

func TestStartClusterUpFlagsPassedThrough(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-start-")
	if err != nil {
		t.Error(err)
	}

	SetMinishiftDir(testDir)
	defer os.RemoveAll(testDir)

	client := http.DefaultClient
	client.Transport = minitesting.NewMockRoundTripper()
	defer minitesting.ResetDefaultRoundTripper()

	testRunner := &RecordingRunner{}
	SetRunner(testRunner)
	defer SetRunner(util.RealRunner{})

	clusterUp()

	expectedOc := filepath.Join(testDir, "cache", "oc", "v1.3.1", "oc")
	if testRunner.Cmd != expectedOc {
		t.Errorf("Expected command '%s'. Got '%s'", expectedOc, testRunner.Cmd)
	}
}
