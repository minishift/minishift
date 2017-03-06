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
