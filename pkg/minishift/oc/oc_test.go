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
)

func Test_invalid_oc_path_returns_error(t *testing.T) {
	invalidPath := "/snafu"
	_, err := NewOcRunner(invalidPath, "")
	if err == nil {
		t.Fatal("An error should have been returned for creating on OcRunner against an invalid path")
	}

	expectedError := fmt.Sprintf(invalidOcPathError, invalidPath)
	if err.Error() != expectedError {
		t.Fatal(fmt.Sprintf("Wrong error returns. Expected '%s'. Got '%s'", expectedError, err.Error()))
	}
}

func Test_invalid_kube_path_returns_error(t *testing.T) {
	// for now it is enough to just pass a file, there are no checks for name of the file or whether
	// it is executable
	tmpOc, err := ioutil.TempFile("", "oc")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpOc.Name()) // clean up

	invalidPath := "/snafu"
	_, err = NewOcRunner(tmpOc.Name(), invalidPath)
	if err == nil {
		t.Fatal("An error should have been returned for creating on OcRunner against an invalid path")
	}

	expectedError := fmt.Sprintf(invalidKubeConfigPathError, invalidPath)
	if err.Error() != expectedError {
		t.Fatal(fmt.Sprintf("Wrong error returns. Expected '%s'. Got '%s'", expectedError, err.Error()))
	}
}

func Test_quotes_in_commands_get_properly_parsed(t *testing.T) {
	fakeRunner := FakeRunner{}
	expectedCommand := "/foo/bar"
	ocRunner := OcRunner{
		ocPath:         expectedCommand,
		kubeConfigPath: "/kube/config",
		runner:         &fakeRunner,
	}

	ocRunner.Run("adm new-project foo --description='Bar Baz'", nil, nil)

	if fakeRunner.cmd != expectedCommand {
		t.Fatal(fmt.Sprintf("Wrong command. Expected '%s'. Got '%s'", expectedCommand, fakeRunner.cmd))
	}

	expectedLength := 5
	if len(fakeRunner.args) != expectedLength {
		t.Fatal(fmt.Sprintf("Wrong args count. Expected '%d'. Got '%d'", expectedLength, len(fakeRunner.args)))
	}

	expectedArgs := []string{"--config=/kube/config", "adm", "new-project", "foo", "--description='Bar Baz'"}
	for i, expectedArg := range expectedArgs {
		if expectedArg != fakeRunner.args[i] {
			t.Error(fmt.Sprintf("Wrong argument. Expected '%s'. Got '%s'", expectedArg, fakeRunner.args[i]))
		}
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
