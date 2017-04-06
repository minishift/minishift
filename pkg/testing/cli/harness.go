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
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/viper"
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

// CaptureStreamOut creates a pipe to capture standard output/Error and returns the original handle to stdout/stderr as well
// as a file handles for the created pipe depend on expectedExitCode.
func CaptureStreamOut(t *testing.T, expectedExitCode int) (*os.File, *os.File, *os.File, *os.File) {
	origStdout := os.Stdout
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	if expectedExitCode == 0 {
		os.Stdout = w
	} else {
		os.Stderr = w
	}
	return origStdout, origStderr, w, r
}

func CreateExitHandlerFunc(t *testing.T, streamWriter *os.File, streamReader *os.File, expectedExitCode int, expectedErrorMessage string) func(int) int {
	var exitHandler func(int) int
	exitHandler = func(code int) int {
		streamWriter.Close()

		var buffer bytes.Buffer
		io.Copy(&buffer, streamReader)

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

func TearDown(testDir string, origStdout *os.File, origStderr *os.File) {
	os.RemoveAll(testDir)
	resetOriginalStdoutFileHandle(origStdout)
	resetOriginalStderrFileHandle(origStderr)
	viper.Reset()
	atexit.ClearExitHandler()
}

func resetOriginalStdoutFileHandle(origStdout *os.File) {
	os.Stdout = origStdout
}

func resetOriginalStderrFileHandle(origStderr *os.File) {
	os.Stderr = origStderr
}
