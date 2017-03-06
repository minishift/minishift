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
	"bytes"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// SetupTmpMinishiftHome creates a tmp directory and points MINISHIFT_HOME to it.
// It returns the path to this tmp directory
func SetupTmpMinishiftHome(t *testing.T) string {
	var err error
	testDir, err := ioutil.TempDir("", "minishift-test-addon-install-cmd-")
	if err != nil {
		t.Error(err)
	}
	constants.Minipath = testDir

	return testDir
}

// CaptureStdOut creates a pipe to capture standard output and returns the original handle to stdout as well
// as a file handles for the created pipe.
func CaptureStdOut(t *testing.T) (*os.File, *os.File, *os.File) {
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stdout = w

	return origStdout, w, r
}

func CreateExitHandlerFunc(t *testing.T, stdOutWriter *os.File, stdOutReader *os.File, expectedExitCode int, expectedErrorMessage string) func(int) int {
	var exitHandler func(int) int
	exitHandler = func(code int) int {
		stdOutWriter.Close()

		var buffer bytes.Buffer
		io.Copy(&buffer, stdOutReader)

		if !strings.HasPrefix(buffer.String(), expectedErrorMessage) {
			t.Fatalf("Expected error '%s'. Got '%s'.", expectedErrorMessage, buffer.String())
		}

		if code != expectedExitCode {
			t.Fatalf("Expected exit code %d. Got %d.", expectedExitCode, code)
		}

		return 0
	}
	return exitHandler
}

func TearDown(testDir string, origStdout *os.File) {
	os.RemoveAll(testDir)
	resetOriginalFileHandle(origStdout)
	viper.Reset()
	atexit.ClearExitHandler()
}

func resetOriginalFileHandle(origStdout *os.File) {
	os.Stdout = origStdout
}
