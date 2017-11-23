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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-github/github"
	minitesting "github.com/minishift/minishift/pkg/testing"
	minishiftos "github.com/minishift/minishift/pkg/util/os"
	"github.com/stretchr/testify/assert"
)

var (
	_, b, _, _   = runtime.Caller(0)
	basepath     = filepath.Dir(b)
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
	{OC, minishiftos.LINUX, testVersion, 2489310, "openshift-origin-client-tools-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-linux-64bit.tar.gz"},
	{OC, minishiftos.DARWIN, testVersion, 2586147, "openshift-origin-client-tools-v1.3.1-2748423-mac.zip"},
	{OC, minishiftos.WINDOWS, testVersion, 2489312, "openshift-origin-client-tools-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-windows.zip"},
}

func TestGetAssetIdAndFilename(t *testing.T) {
	EnsureGitHubApiAccessTokenSet(t)

	for _, testAsset := range assetSet {
		release, resp, err = gitHubClient.Repositories.GetReleaseByTag("openshift", "origin", testAsset.version)
		assert.NoError(t, err, "Could not get OpenShift release")
		defer func() {
			_ = resp.Body.Close()
		}()

		actualAssetID, actualFilename := getAssetIdAndFilename(testAsset.binary, testAsset.os, release)
		assert.Equal(t, testAsset.expectedAssetId, actualAssetID, "Unexpected asset id for binary %s for OS %s",
			testAsset.binary, testAsset.os)

		assert.Equal(t, testAsset.expectedFilename, actualFilename, "Unexpected filename for binary %s for OS %s.",
			testAsset.binary, testAsset.os)
	}
}

func TestDownloadOc(t *testing.T) {
	EnsureGitHubApiAccessTokenSet(t)

	mockTransport := minitesting.NewMockRoundTripper()
	addMockResponses(mockTransport)

	client := http.DefaultClient
	client.Transport = mockTransport

	defer minitesting.ResetDefaultRoundTripper()

	testDir, err := ioutil.TempDir("", "minishift-test-")
	assert.NoError(t, err, "Error crating temp directory")
	defer os.RemoveAll(testDir)

	for _, testAsset := range assetSet {
		err = DownloadOpenShiftReleaseBinary(testAsset.binary, testAsset.os, testAsset.version, testDir)
		assert.NoError(t, err, "Error in downloading OpenShift release binary")

		expectedBinaryPath := filepath.Join(testDir, testAsset.binary.String())
		if testAsset.os == minishiftos.WINDOWS {
			expectedBinaryPath += ".exe"
		}
		fileInfo, err := os.Lstat(expectedBinaryPath)
		assert.NoError(t, err, "Error in getting fileinfo")

		if runtime.GOOS != "windows" {
			expectedFilePermissions := "-rwxrwxrwx"
			assert.Equal(t, expectedFilePermissions, fileInfo.Mode().String())
		}

		err = os.Remove(expectedBinaryPath)
		assert.NoError(t, err, "Error in removing expected binary path")
	}
}

func TestInvalidVersion(t *testing.T) {
	EnsureGitHubApiAccessTokenSet(t)

	testDir, err := ioutil.TempDir("", "minishift-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(testDir)

	dummyVersion := "foo"
	err = DownloadOpenShiftReleaseBinary(OPENSHIFT, minishiftos.WINDOWS, dummyVersion, testDir)
	assert.Error(t, err, "Error in downloading OpenShift release binary")

	expectedErrorMessage := fmt.Sprintf("Cannot get the OpenShift release version %s: GET https://api.github.com/repos/openshift/origin/releases/tags/foo: 404 Not Found []", dummyVersion)
	assert.EqualError(t, err, expectedErrorMessage)
}

func TestInvalidBinaryFormat(t *testing.T) {
	EnsureGitHubApiAccessTokenSet(t)

	testDir, err := ioutil.TempDir("", "minishift-test-")
	assert.NoError(t, err, "Error creating temp directory")
	defer os.RemoveAll(testDir)

	err = DownloadOpenShiftReleaseBinary(OPENSHIFT, minishiftos.WINDOWS, testVersion, testDir)
	assert.Error(t, err, "Error in downloading OpenShift release binary")

	expectedErrorMessage := "Cannot get binary 'openshift' in version v1.3.1 for the target environment Windows"
	assert.EqualError(t, err, expectedErrorMessage)
}

// See https://github.com/minishift/minishift/issues/331
func Test_Download_Oc_1_4_1(t *testing.T) {
	EnsureGitHubApiAccessTokenSet(t)

	mockTransport := minitesting.NewMockRoundTripper()
	addMockResponses(mockTransport)

	client := http.DefaultClient
	client.Transport = mockTransport
	defer minitesting.ResetDefaultRoundTripper()

	testDir, err := ioutil.TempDir("", "minishift-test-")
	assert.NoError(t, err, "Error creating temp directory")
	defer os.RemoveAll(testDir)

	err = DownloadOpenShiftReleaseBinary(OC, minishiftos.LINUX, "v1.4.1", testDir)
	assert.NoError(t, err, "Error in downloading OpenShift binary")

	expectedBinaryPath := filepath.Join(testDir, "oc")

	fileInfo, err := os.Lstat(expectedBinaryPath)
	assert.NoError(t, err, "Error in getting fileinfo")

	if runtime.GOOS != "windows" {
		expectedFilePermissions := "-rwxrwxrwx"
		assert.Equal(t, expectedFilePermissions, fileInfo.Mode().String())
	}

	err = os.Remove(expectedBinaryPath)
	assert.NoError(t, err, "Error in removing expected binary path")
}

func EnsureGitHubApiAccessTokenSet(t *testing.T) {
	if GetGitHubApiToken() == "" {
		t.Skip("Skipping GitHub API based test, because no access token is defined in the environment.\n " +
			"To run this test check https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/ and set for example MINISHIFT_GITHUB_API_TOKEN (see github.go).")
	}
}

func addMockResponses(mockTransport *minitesting.MockRoundTripper) {
	testDataDir := filepath.Join(basepath, "..", "..", "..", "test", "testdata")

	mockTransport.RegisterResponse("https://.*CHECKSUM", &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "CHECKSUM"),
		ContentType:  minitesting.TEXT,
	})

	url := "https://.*openshift-origin-client-tools-v1.3.1-2748423-mac.zip"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "openshift-origin-client-tools-v1.3.1-2748423-mac.zip"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://.*openshift-origin-client-tools-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-linux-64bit.tar.gz"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "openshift-origin-client-tools-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-linux-64bit.tar.gz"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://.*openshift-origin-client-tools-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-windows.zip"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "openshift-origin-client-tools-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-windows.zip"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://.*openshift-origin-client-tools-v1.4.1-3f9807a-linux-64bit.tar.gz"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "openshift-origin-client-tools-v1.4.1-3f9807a-linux-64bit.tar.gz"),
		ContentType:  minitesting.OCTET_STREAM,
	})
}
