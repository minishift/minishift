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
	"github.com/minishift/minishift/cmd/minishift/state"
	pkgTesting "github.com/minishift/minishift/pkg/testing"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/viper"
)

// SetupTmpMinishiftHome creates a tmp directory and points MINISHIFT_HOME to it.
// It returns the path to this tmp directory
func SetupTmpMinishiftHome(t *testing.T) string {
	var err error
	tmpDir, err := ioutil.TempDir("", "minishift-tmp-test-dir-")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("MINISHIFT_HOME", tmpDir)
	state.InstanceDirs = state.NewMinishiftDirs(tmpDir)

	return tmpDir
}

// CreateTee splits the stdout and stderr in order to capture these streams into a buffer
// during test execution. If silent is true, the original output streams are silenced.
func CreateTee(t *testing.T, silent bool) *pkgTesting.Tee {
	tee, err := pkgTesting.NewTee(silent)
	if err != nil {
		t.Fatalf("Unexpected error during setup: %s", err.Error())
	}
	return tee
}

// VerifyExitCodeAndMessage creates an exit handler which verifies that the program will try to exit execution with the specified
// exit code and message.
func VerifyExitCodeAndMessage(t *testing.T, tee *pkgTesting.Tee, expectedExitCode int, expectedErrorMessage string) func(int) bool {
	exitHandler := func(code int) bool {
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

// PreventExitWithNonZeroReturn creates an exit handler function which will cast a veto when the program tries to call os.Exit with a non
// zero exit code. This is useful to prevent a test from exiting early due to failing validation check.
func PreventExitWithNonZeroExitCode(t *testing.T) func(int) bool {
	exitHandler := func(code int) bool {
		return code != 0
	}
	return exitHandler
}

// PreventAtExit prevents an early/unexpected termination via atexit.
func PreventAtExit(t *testing.T) func(int) bool {
	exitHandler := func(code int) bool {
		t.Fatal("The called method unexpectedly called atexit.")
		return true
	}
	return exitHandler
}

func TearDown(testDir string, tee *pkgTesting.Tee) {
	os.Unsetenv("MINISHIFT_HOME")
	if tee != nil {
		tee.Close()
	}
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

// PrepareStdinResponse creates a temproary file with a prepared content which then is used as os.Stdin to test user input.
// The orignal os.Stdin file handle is returned as well as the path of the file containing the canned stdin responses.
func PrepareStdinResponse(s string, t *testing.T) (*os.File, string) {
	content := []byte(s)
	tmpfile, err := ioutil.TempFile("", "minishift-test-input")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	origStdin := os.Stdin
	os.Stdin = tmpfile

	return origStdin, tmpfile.Name()

}

// ResetStdin resets os.Stdin to the specified original file handle. It also deletes a potenitally created tmp file containing dummy stdin data.
func ResetStdin(origStdin *os.File, tmpFile string) {
	os.Stdin = origStdin
	if tmpFile != "" {
		os.Remove(tmpFile)
	}
}
