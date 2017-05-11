/*
Copyright (C) 2017 Red Hat, Inc.

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

package cli

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"fmt"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/viper"
)

// SetupTmpMinishiftHome creates a tmp directory and points MINISHIFT_HOME to it.
// It returns the path to this tmp directory
func SetupTmpMinishiftHome(t *testing.T) string {
	var err error
	testDir, err := ioutil.TempDir("", "minishift-tmp-test-dir-")
	if err != nil {
		t.Fatal(err)
	}
	constants.Minipath = testDir

	return testDir
}

// CreateTee splits the stdout and stderr in order to capture these streams into a buffer
// during test execution. If silent is true, the original output streams are silenced.
func CreateTee(t *testing.T, silent bool) *Tee {
	tee, err := NewTee(silent)
	if err != nil {
		t.Fatalf("Unexpected error during setup: %s", err.Error())
	}
	return tee
}

func CreateExitHandlerFunc(t *testing.T, tee *Tee, expectedExitCode int, expectedErrorMessage string) func(int) bool {
	var exitHandler func(int) bool
	exitHandler = func(code int) bool {
		tee.Close()

		var actualOutput string
		if code == 0 {
			actualOutput = tee.StdoutBuffer.String()
		} else {
			actualOutput = tee.StderrBuffer.String()
		}

		if !strings.HasPrefix(actualOutput, expectedErrorMessage) {
			t.Fatalf("Expected error '%s'. Got '%s'.", expectedErrorMessage, tee.StderrBuffer.String())
		}

		if code != expectedExitCode {
			t.Fatalf("Expected exit code %d. Got %d.", expectedExitCode, code)
		}

		return true
	}
	return exitHandler
}

func TearDown(testDir string, tee *Tee) {
	tee.Close()
	os.RemoveAll(testDir)
	viper.Reset()
	atexit.ClearExitHandler()
	if r := recover(); r != nil {
		reason := fmt.Sprint(r)
		if reason != atexit.ExitHandlerPanicMessage {
			fmt.Println("Recovered from panic:", r)
		}
	}
}
