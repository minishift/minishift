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

package openshift

import (
	"testing"

	"bytes"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/util"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

var testDir string
var r *os.File
var w *os.File
var origStdout *os.File

func Test_unknown_patch_target_aborts_command(t *testing.T) {
	setup(t)
	defer tearDown()

	util.RegisterExitHandler(createExitHandlerFunc(t, 1, unknownPatchTargetError))

	target = "foo"
	runPatch(nil, nil)
}

func Test_patch_cannot_be_empty(t *testing.T) {
	setup(t)
	defer tearDown()

	util.RegisterExitHandler(createExitHandlerFunc(t, 1, emptyPatchError))

	target = "master"
	patch = ""
	runPatch(nil, nil)
}

func Test_patch_needs_to_be_valid_JSON(t *testing.T) {
	setup(t)
	defer tearDown()

	util.RegisterExitHandler(createExitHandlerFunc(t, 1, invalidJSONError))

	target = "master"
	patch = "foo"
	runPatch(nil, nil)
}

func Test_patch_commands_needs_existing_vm(t *testing.T) {
	setup(t)
	defer tearDown()

	util.RegisterExitHandler(createExitHandlerFunc(t, 1, nonExistentMachineError))

	target = "master"
	patch = "{\"corsAllowedOrigins\": \"*\"}"
	runPatch(nil, nil)
}

func resetOriginalFileHandle() {
	os.Stdout = origStdout
}

func setup(t *testing.T) {
	var err error
	testDir, err = ioutil.TempDir("", "minishift-test-patch-cmd-")
	if err != nil {
		t.Error(err)
	}

	origStdout = os.Stdout
	r, w, err = os.Pipe()
	if err != nil {
		t.Error(err)
	}
	os.Stdout = w

	constants.Minipath = testDir
}

func tearDown() {
	os.RemoveAll(testDir)
	resetOriginalFileHandle()
	viper.Reset()
	util.ClearExitHandler()
}

func createExitHandlerFunc(t *testing.T, expectedExitCode int, expectedErrorMessage string) func(int) int {
	var exitHandler func(int) int
	exitHandler = func(code int) int {
		w.Close()
		var buffer bytes.Buffer
		io.Copy(&buffer, r)

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
