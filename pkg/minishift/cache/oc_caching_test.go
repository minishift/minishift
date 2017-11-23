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

package cache

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/minishift/minishift/pkg/minikube/constants"
	minitesting "github.com/minishift/minishift/pkg/testing"
	"github.com/stretchr/testify/assert"
)

var (
	testDir string
	testOc  Oc
)

func TestIsCached(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)

	ocDir := filepath.Join(testDir, "cache", "oc", "v1.3.1", runtime.GOOS)
	os.MkdirAll(ocDir, os.ModePerm)

	assert.False(t, testOc.isCached())

	content := []byte("foo")

	err := ioutil.WriteFile(filepath.Join(ocDir, constants.OC_BINARY_NAME), content, os.ModePerm)
	assert.NoError(t, err, "Error writing to file")

	assert.True(t, testOc.isCached())
}

func TestCacheOc(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir) // clean up

	mockTransport := minitesting.NewMockRoundTripper()
	addMockResponses(mockTransport)

	client := http.DefaultClient
	client.Transport = mockTransport

	defer minitesting.ResetDefaultRoundTripper()

	ocDir := filepath.Join(testDir, "cache", "oc", "v1.3.1")
	os.MkdirAll(ocDir, os.ModePerm)

	err := testOc.cacheOc()
	assert.NoError(t, err, "Error caching oc")
}

func setUp(t *testing.T) {
	var err error
	testDir, err = ioutil.TempDir("", "minishift-test-")
	if err != nil {
		t.Error()
	}
	testOc = Oc{"v1.3.1", filepath.Join(testDir, "cache")}
}

func addMockResponses(mockTransport *minitesting.MockRoundTripper) {
	testDataDir := filepath.Join("..", "..", "..", "test", "testdata")

	mockTransport.RegisterResponse("https://api.github.com/repos/openshift/origin/releases/tags/v1.3.1", &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "openshift-1.3.1-release-tag.json"),
		ContentType:  minitesting.JSON,
	})

	var assetContent string
	switch runtime.GOOS {
	case "windows":
		assetContent = filepath.Join(testDataDir, "openshift-origin-client-tools-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-windows.zip")
		mockTransport.RegisterResponse("https://api.github.com/repos/openshift/origin/releases/assets/2489312", &minitesting.CannedResponse{
			ResponseType: minitesting.SERVE_FILE,
			Response:     assetContent,
			ContentType:  minitesting.OCTET_STREAM,
		})
	case "darwin":
		assetContent = filepath.Join(testDataDir, "openshift-origin-client-tools-v1.3.1-2748423-mac.zip")
		mockTransport.RegisterResponse("https://api.github.com/repos/openshift/origin/releases/assets/2586147", &minitesting.CannedResponse{
			ResponseType: minitesting.SERVE_FILE,
			Response:     assetContent,
			ContentType:  minitesting.OCTET_STREAM,
		})
	case "linux":
		assetContent = filepath.Join(testDataDir, "openshift-origin-client-tools-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-linux-64bit.tar.gz")
		mockTransport.RegisterResponse("https://api.github.com/repos/openshift/origin/releases/assets/2489310", &minitesting.CannedResponse{
			ResponseType: minitesting.SERVE_FILE,
			Response:     assetContent,
			ContentType:  minitesting.OCTET_STREAM,
		})
	}

	mockTransport.RegisterResponse("https://api.github.com/repos/openshift/origin/releases/assets/2489308", &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "CHECKSUM"),
		ContentType:  minitesting.TEXT,
	})
}
