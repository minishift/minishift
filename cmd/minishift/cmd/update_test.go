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

package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

func TestCreateUpdateMarker(t *testing.T) {
	testFile, err := ioutil.TempFile(os.TempDir(), "testFile")
	testFileName := testFile.Name()
	defer os.Remove(testFileName)
	if err != nil {
		t.Fatal("Unexpected error: " + err.Error())
	}

	expectedInstallAddon, expectedPreviousVersion := true, "1.0.0"
	createUpdateMarker(testFileName, UpdateMarker{expectedInstallAddon, expectedPreviousVersion})

	// Read json file
	var markerData UpdateMarker

	file, err := ioutil.ReadFile(testFileName)
	if err != nil {
		t.Fatal("Unexpected error: " + err.Error())
	}

	json.Unmarshal(file, &markerData)
	if markerData.InstallAddon != expectedInstallAddon {
		t.Fatalf("Expected allow installation of addon to be %t but got %t.", expectedInstallAddon, markerData.InstallAddon)
	}
	if markerData.PreviousVersion != expectedPreviousVersion {
		t.Fatalf("Expected previous version to be %t but got %t.", expectedPreviousVersion, markerData.PreviousVersion)
	}
}
