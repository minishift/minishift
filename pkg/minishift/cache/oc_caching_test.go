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
	minitesting "github.com/minishift/minishift/pkg/testing"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"github.com/minishift/minishift/pkg/minikube/constants"
)

var testDir string
var testOc Oc

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

	client := http.DefaultClient
	client.Transport = minitesting.NewMockRoundTripper()
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
