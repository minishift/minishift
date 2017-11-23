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

package update

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	minitesting "github.com/minishift/minishift/pkg/testing"
	minishiftos "github.com/minishift/minishift/pkg/util/os"
	"github.com/stretchr/testify/assert"
)

var (
	_, b, _, _          = runtime.Caller(0)
	basepath            = filepath.Dir(b)
	err                 error
	testDir             string
	expectedArchivePath string
	testDataDir         string
	assetSet            = []struct {
		os          minishiftos.OS
		version     string
		archiveName string
	}{
		{minishiftos.LINUX, "1.0.0", "minishift-1.0.0-linux-amd64.tgz"},
		{minishiftos.WINDOWS, "1.0.0", "minishift-1.0.0-windows-amd64.zip"},
		{minishiftos.LINUX, "1.2.0", "minishift-1.2.0-linux-amd64.tgz"},
		{minishiftos.WINDOWS, "1.2.0", "minishift-1.2.0-windows-amd64.zip"},
		{minishiftos.LINUX, "1.2.1", "minishift-1.2.1-linux-amd64.tgz"},
		{minishiftos.WINDOWS, "1.2.1", "minishift-1.2.1-windows-amd64.zip"},
	}
)

func TestDownloadAndVerifyArchive(t *testing.T) {
	mockTransport := minitesting.NewMockRoundTripper()
	addMockResponses(mockTransport)

	client := http.DefaultClient
	client.Transport = mockTransport
	defer minitesting.ResetDefaultRoundTripper()

	setUp(t)
	defer os.RemoveAll(testDir)

	for _, testAsset := range assetSet {
		expectedArchivePath = filepath.Join(testDir, testAsset.archiveName)
		downloadLinkFormat := "https://github.com/" + githubOwner + "/" + githubRepo + "/releases/download/v%s/%s"
		url := fmt.Sprintf(downloadLinkFormat, testAsset.version, testAsset.archiveName)
		archivePath, err := downloadAndVerifyArchive(url, testDir)

		assert.NoError(t, err, "Error in downloading and verifying archive")
		assert.Equal(t, expectedArchivePath, archivePath)
	}
}

func TestExtractBinaryforOlderVersionFormat(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	testDataDir = filepath.Join(basepath, "..", "..", "..", "test", "testdata")

	extName := "tgz"
	osName := "linux"
	binaryName := "minishift"
	if runtime.GOOS == "windows" {
		extName = "zip"
		osName = "windows"
		binaryName = "minishift.exe"
	}
	archiveName := fmt.Sprintf("minishift-%s-%s-%s.%s", "1.0.0", osName, "amd64", extName)
	testArchivePath := filepath.Join(testDataDir, archiveName)
	tmpArchivePath := filepath.Join(testDir, archiveName)
	// Copy test archive to tmpDir for testing purpose
	copyArchive(t, testArchivePath, tmpArchivePath)
	fileName, err := extractBinary(tmpArchivePath, testDir)
	assert.NoError(t, err, "Error in extracting binary")
	assert.Contains(t, fileName, binaryName)
}

func TestExtractBinaryforNewerVersionFormat(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	testDataDir = filepath.Join(basepath, "..", "..", "..", "test", "testdata")

	extName := "tgz"
	osName := "linux"
	binaryName := "minishift"
	if runtime.GOOS == "windows" {
		extName = "zip"
		osName = "windows"
		binaryName = "minishift.exe"
	}
	archiveName := fmt.Sprintf("minishift-%s-%s-%s.%s", "1.2.0", osName, "amd64", extName)
	testArchivePath := filepath.Join(testDataDir, archiveName)
	tmpArchivePath := filepath.Join(testDir, archiveName)
	// Copy test archive to tmpDir for testing purpose
	copyArchive(t, testArchivePath, tmpArchivePath)
	fileName, err := extractBinary(tmpArchivePath, testDir)
	assert.NoError(t, err, "Error in extracting binary")
	assert.Contains(t, fileName, binaryName)

}

func TestExtractBinaryWithoutHavingMiniShift(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	testDataDir = filepath.Join(basepath, "..", "..", "..", "test", "testdata")

	extName := "tgz"
	osName := "linux"
	if runtime.GOOS == "windows" {
		extName = "zip"
		osName = "windows"
	}
	archiveName := fmt.Sprintf("minishift-%s-%s-%s.%s", "1.2.1", osName, "amd64", extName)
	testArchivePath := filepath.Join(testDataDir, archiveName)
	tmpArchivePath := filepath.Join(testDir, archiveName)
	// Copy test archive to tmpDir for testing purpose
	copyArchive(t, testArchivePath, tmpArchivePath)
	fileName, err := extractBinary(tmpArchivePath, testDir)
	assert.NoError(t, err, "Error in extracting binary")
	assert.Empty(t, fileName)

}

func setUp(t *testing.T) {
	testDir, err = ioutil.TempDir("", "minishift-test-")
	assert.NoError(t, err)
}

func copyArchive(t *testing.T, src, dest string) {
	data, err := ioutil.ReadFile(src)
	assert.NoError(t, err)
	err = ioutil.WriteFile(dest, data, 0644)
	assert.NoError(t, err, "Error writing to file")
}

func addMockResponses(mockTransport *minitesting.MockRoundTripper) {
	testDataDir := filepath.Join(basepath, "..", "..", "..", "test", "testdata")

	url := "https://github.com/minishift/minishift/releases/download/v1.0.0/minishift-1.0.0-windows-amd64.zip$"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "minishift-1.0.0-windows-amd64.zip"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://github.com/minishift/minishift/releases/download/v1.0.0/minishift-1.0.0-windows-amd64.zip.sha256"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "minishift-1.0.0-windows-amd64.zip.sha256"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://github.com/minishift/minishift/releases/download/v1.0.0/minishift-1.0.0-linux-amd64.tgz$"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "minishift-1.0.0-linux-amd64.tgz"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://github.com/minishift/minishift/releases/download/v1.0.0/minishift-1.0.0-linux-amd64.tgz.sha256"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "minishift-1.0.0-linux-amd64.tgz.sha256"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://github.com/minishift/minishift/releases/download/v1.2.0/minishift-1.2.0-windows-amd64.zip$"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "minishift-1.2.0-windows-amd64.zip"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://github.com/minishift/minishift/releases/download/v1.2.0/minishift-1.2.0-windows-amd64.zip.sha256"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "minishift-1.2.0-windows-amd64.zip.sha256"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://github.com/minishift/minishift/releases/download/v1.2.0/minishift-1.2.0-linux-amd64.tgz$"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "minishift-1.2.0-linux-amd64.tgz"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://github.com/minishift/minishift/releases/download/v1.2.0/minishift-1.2.0-linux-amd64.tgz.sha256"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "minishift-1.2.0-linux-amd64.tgz.sha256"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://github.com/minishift/minishift/releases/download/v1.2.1/minishift-1.2.1-windows-amd64.zip$"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "minishift-1.2.1-windows-amd64.zip"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://github.com/minishift/minishift/releases/download/v1.2.1/minishift-1.2.1-windows-amd64.zip.sha256"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "minishift-1.2.1-windows-amd64.zip.sha256"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://github.com/minishift/minishift/releases/download/v1.2.1/minishift-1.2.1-linux-amd64.tgz$"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "minishift-1.2.1-linux-amd64.tgz"),
		ContentType:  minitesting.OCTET_STREAM,
	})

	url = "https://github.com/minishift/minishift/releases/download/v1.2.1/minishift-1.2.1-linux-amd64.tgz.sha256"
	mockTransport.RegisterResponse(url, &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "minishift-1.2.1-linux-amd64.tgz.sha256"),
		ContentType:  minitesting.OCTET_STREAM,
	})
}
