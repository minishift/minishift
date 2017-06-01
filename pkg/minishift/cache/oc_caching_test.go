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
	"github.com/minishift/minishift/pkg/minikube/constants"
	minitesting "github.com/minishift/minishift/pkg/testing"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
	testDir    string
	testOc     Oc
)

func TestIsCached(t *testing.T) {
	setUp(t)
	defer os.RemoveAll(testDir)

	ocDir := filepath.Join(testDir, "cache", "oc", "v1.3.1")
	os.MkdirAll(ocDir, os.ModePerm)

	if testOc.isCached() != false {
		t.Error("Expected oc to be uncached")
	}

	content := []byte("foo")

	if err := ioutil.WriteFile(filepath.Join(ocDir, constants.OC_BINARY_NAME), content, os.ModePerm); err != nil {
		t.Error()
	}

	if testOc.isCached() != true {
		t.Error("Expected oc to be cached")
	}
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
	if err != nil {
		t.Error(err)
	}
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
	testDataDir := filepath.Join(basepath, "..", "..", "..", "test", "testdata")

	mockTransport.RegisterResponse("https://api.github.com/repos/openshift/origin/releases/tags/v1.3.1", &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     filepath.Join(testDataDir, "openshift-1.3.1-release-tag.json"),
		ContentType:  minitesting.JSON,
	})

	var assetContent string
	switch runtime.GOOS {
	case "windows":
		assetContent = filepath.Join(testDataDir, "openshift-origin-client-tools-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-windows.zip")
	case "darwin":
		assetContent = filepath.Join(testDataDir, "openshift-origin-client-tools-v1.3.1-2748423-mac.zip")
	case "linux":
		assetContent = filepath.Join(testDataDir, "openshift-origin-client-tools-v1.3.1-dad658de7465ba8a234a4fb40b5b446a45a4cee1-linux-64bit.tar.gz")
	}
	mockTransport.RegisterResponse("https://api.github.com/repos/openshift/origin/releases/assets/.*", &minitesting.CannedResponse{
		ResponseType: minitesting.SERVE_FILE,
		Response:     assetContent,
		ContentType:  minitesting.OCTET_STREAM,
	})
}
