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

package clusterup

import (
	"github.com/pkg/errors"
	"testing"
)

func TestValidateOpenshiftMinVersion(t *testing.T) {
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
		valid, err := ValidateOpenshiftMinVersion(versionTest.version, minVer)
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
