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

package version

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/version"
)

func TestPrintUpStreamVersions(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-config-")
	if err != nil {
		t.Error()
	}
	defer os.RemoveAll(testDir)

	f, err := os.Create(testDir + "out.txt")
	if err != nil {
		t.Fatal("Error creating test file", err)
	}
	defer f.Close()

	os.Stdout = f
	defaultVersion := version.GetOpenShiftVersion()
	err = PrintUpStreamVersions(f, constants.MinimumSupportedOpenShiftVersion, defaultVersion)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal("Error setting offset back", err)
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal("Error reading file", err)
	}
	actualStdout := string(data)
	if !strings.Contains(actualStdout, constants.MinimumSupportedOpenShiftVersion) {
		t.Fatalf("Should Contain constants.MinimumSupportedOpenShiftVersion in\n %s", actualStdout)
	}
	if !strings.Contains(actualStdout, defaultVersion) {
		t.Fatalf("Should Contain defaultVersion in\n %s", actualStdout)
	}
}

func TestPrintDownStreamVersions(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-config-")
	if err != nil {
		t.Error()
	}
	defer os.RemoveAll(testDir)

	f, err := os.Create(testDir + "out.txt")
	if err != nil {
		t.Fatal("Error creating test file", err)
	}
	defer f.Close()

	os.Stdout = f
	err = PrintDownStreamVersions(f, "v3.4.1.10")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal("Error setting offset back", err)
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal("Error reading file", err)
	}

	actualStdout := string(data)
	if !strings.Contains(actualStdout, "v3.4.1.10") {
		t.Fatalf("Should Contain v3.4.1.10 in\n %s", actualStdout)
	}
}

func Test_validate_openshift_min_versions(t *testing.T) {
	var versionTests = []struct {
		version string // input
		valid   bool   // expected result
		err     error
	}{
		{"v1.1.0", false, nil},
		{"v1.2.2", false, nil},
		{"v1.2.3-beta", false, nil},
		{"v1.3.1", false, nil},
		{"v1.3.5-alpha", false, nil},
		{"foo", false, errors.New("Invalid version format 'foo': No Major.Minor.Patch elements found")},
		{"151", false, errors.New("Invalid version format '151': No Major.Minor.Patch elements found")},
		{"v1.4.1", true, nil},
		{"v1.5.0-alpha.0", true, nil},
		{"v1.5.1-beta.0", true, nil},
		{"v3.6.0", true, nil},
		{"3.6.0", true, nil},
	}

	minVer := "v1.4.1"
	for _, versionTest := range versionTests {
		valid, err := IsGreaterOrEqualToBaseVersion(versionTest.version, minVer)
		if versionTest.err == nil && err != nil {
			t.Fatalf("No error expected. Got '%v'", err)
		}

		if err != nil && err.Error() != versionTest.err.Error() {
			t.Fatalf("Unexpected error. Expected '%v', got '%v'", versionTest.err, err)
		}

		if valid != versionTest.valid {
			t.Fatalf("Expected '%t' Got '%t' for %s", versionTest.valid, valid, versionTest.version)
		}
	}
}
