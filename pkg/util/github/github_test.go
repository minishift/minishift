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

package github

import (
	"github.com/google/go-github/github"
	minitesting "github.com/minishift/minishift/pkg/testing"
	minishiftos "github.com/minishift/minishift/pkg/util/os"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"fmt"
	"strings"
	"runtime"
)

var (
	gitHubClient = Client()
	err          error
	release      *github.RepositoryRelease
	resp         *github.Response
)

var testVersion = "v1.3.1"
var assetSet = []struct {
	binary           OpenShiftBinaryType
	os               minishiftos.OS
	version          string
	expectedAssetId  int
	expectedFilename string
}{
	{OPENSHIFT, minishiftos.LINUX, testVersion, 2489309, "openshift-origin-server-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-linux-64bit.tar.gz"},
	{OC, minishiftos.LINUX, testVersion, 2489310, "openshift-origin-client-tools-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-linux-64bit.tar.gz"},
	{OC, minishiftos.DARWIN, testVersion, 2586147, "openshift-origin-client-tools-v1.3.1-2748423-mac.zip"},
	{OC, minishiftos.WINDOWS, testVersion, 2489312, "openshift-origin-client-tools-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-windows.zip"},
}

func TestGetAssetIdAndFilename(t *testing.T) {
	for _, testAsset := range assetSet {
		release, resp, err = gitHubClient.Repositories.GetReleaseByTag("openshift", "origin", testAsset.version)
		if err != nil {
			t.Error(err, "Could not get OpenShift release")
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		actualAssetID, actualFilename := getAssetIdAndFilename(testAsset.binary, testAsset.os, release)
		if actualAssetID != testAsset.expectedAssetId {
			t.Errorf("Unexpected asset id for binary %d for OS %d. Expected %d, got %d",
				testAsset.binary, testAsset.os, testAsset.expectedAssetId, actualAssetID)
		}

		if actualFilename != testAsset.expectedFilename {
			t.Errorf("Unexpected filename for binary %d for OS %d. Expected %s, got %s",
				testAsset.binary, testAsset.os, testAsset.expectedFilename, actualFilename)
		}
	}
}

func TestDownloadOc(t *testing.T) {
	client := http.DefaultClient
	client.Transport = minitesting.NewMockRoundTripper()
	defer minitesting.ResetDefaultRoundTripper()

	testDir, err := ioutil.TempDir("", "minishift-test-")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(testDir)

	for _, testAsset := range assetSet {
		// don't test with openshift binary
		if testAsset.binary == OPENSHIFT {
			continue
		}

		err = DownloadOpenShiftReleaseBinary(testAsset.binary, testAsset.os, testAsset.version, testDir)
		if err != nil {
			t.Error(err)
		}

		expectedBinaryPath := filepath.Join(testDir, testAsset.binary.String())
		if testAsset.os == minishiftos.WINDOWS {
			expectedBinaryPath += ".exe"
		}
		fileInfo, err := os.Lstat(expectedBinaryPath)
		if err != nil {
			t.Error(err)
		}

		if runtime.GOOS != "windows" {
			expectedFilePermissions := "-rwxrwxrwx"
			if fileInfo.Mode().String() != expectedFilePermissions {
				t.Errorf("Wrong file permisisons. Expected %s. Got %s", expectedFilePermissions, fileInfo.Mode().String())
			}
		}

		err = os.Remove(expectedBinaryPath)
		if err != nil {
			t.Errorf("Unable to delete %s", expectedBinaryPath)
		}
	}
}

func TestInvalidVersion(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(testDir)

	dummyVersion := "foo"
	err = DownloadOpenShiftReleaseBinary(OPENSHIFT, minishiftos.WINDOWS, dummyVersion, testDir)
	if err == nil {
		t.Error("There should have been an error")
	}

	expectedErrorMessage := fmt.Sprintf("Cannot get the OpenShift release version %s", dummyVersion)
	if !strings.HasPrefix(err.Error(), expectedErrorMessage) {
		t.Errorf("Expected error: '%s'. Got: '%s'\n", expectedErrorMessage, err.Error())
	}
}

func TestInvalidBinaryFormat(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(testDir)

	err = DownloadOpenShiftReleaseBinary(OPENSHIFT, minishiftos.WINDOWS, testVersion, testDir)
	if err == nil {
		t.Error("There should have been an error")
	}

	expectedErrorMessage := "Cannot get binary 'openshift' in version v1.3.1 for the target environment Windows"
	if err.Error() != expectedErrorMessage {
		t.Errorf("Expected error: '%s'. Got: '%s'\n", expectedErrorMessage, err.Error())
	}
}
