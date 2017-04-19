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

package manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/minishift/minishift/pkg/minishift/addon"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

var anyuid string = `# Name: anyuid
# Description: Allows authenticated users to run images to run with USER as per Dockerfile

oc adm policy add-scc-to-group anyuid system:authenticated
`

func Test_creating_addon_manager_for_non_existing_directory_returns_an_error(t *testing.T) {
	path := filepath.Join("this", "path", "really", "should", "not", "exists", "unless", "you", "have", "a", "crazy", "setup")

	_, err := NewAddOnManager(path, make(map[string]*addon.AddOnConfig))

	if err == nil {
		t.Fatal(fmt.Sprintf("Creating the manager in directory '%s' should have failed", path))
	}

	if !strings.HasPrefix(err.Error(), "Unable to create addon manager") {
		t.Fatal(fmt.Sprintf("Unexpected error message '%s", err.Error()))
	}
}

func Test_create_addon_manager(t *testing.T) {
	path := filepath.Join(basepath, "..", "..", "..", "..", "addons")

	manager, err := NewAddOnManager(path, make(map[string]*addon.AddOnConfig))

	if err != nil {
		t.Fatal(fmt.Sprintf("Unexpected error creating manager in directory '%s'. Creation should not have failed: '%s'", path, err.Error()))
	}

	addOns := manager.List()

	expectedNumberOfAddOns := 5
	actualNumberOfAddOns := len(addOns)
	if actualNumberOfAddOns != expectedNumberOfAddOns {
		t.Fatal(fmt.Sprintf("Unexpected number of addons. Expected %d, got %d", expectedNumberOfAddOns, actualNumberOfAddOns))
	}
}

func Test_installing_addon_for_non_existing_directory_returns_an_error(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-addon-manager-")
	defer os.RemoveAll(testDir)

	manager, err := NewAddOnManager(testDir, make(map[string]*addon.AddOnConfig))
	_, err = manager.Install("foo", false)

	if err == nil {
		t.Fatal(("Creating addon should have failed"))
	}

	if !strings.HasPrefix(err.Error(), "The source of a addon needs to be a directory") {
		t.Fatal(fmt.Sprintf("Unexpected error message '%s", err.Error()))
	}
}

func Test_invalid_addons_get_skipped(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-addon-manager-")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(testDir)

	// create a valid addon
	addOn1Path := filepath.Join(testDir, "addon1")
	err = os.Mkdir(addOn1Path, 0777)
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile(filepath.Join(addOn1Path, "anyuid.addon"), []byte(anyuid), 0777)
	if err != nil {
		t.Error(err)
	}

	// create a invalid addon
	addOn2Path := filepath.Join(testDir, "addon2")
	err = os.Mkdir(addOn2Path, 0777)
	if err != nil {
		t.Error(err)
	}
	_, err = os.Create(filepath.Join(addOn2Path, "foo.addon"))
	if err != nil {
		t.Error(err)
	}

	manager, err := NewAddOnManager(testDir, make(map[string]*addon.AddOnConfig))

	if err != nil {
		t.Fatal(fmt.Sprintf("Unexpected error creating manager in directory '%s'. Creation should not have failed: '%s'", testDir, err.Error()))
	}

	addOns := manager.List()

	expectedNumberOfAddOns := 1
	actualNumberOfAddOns := len(addOns)
	if actualNumberOfAddOns != expectedNumberOfAddOns {
		t.Fatal(fmt.Sprintf("Unexpected number of addons. Expected %d, got %d", expectedNumberOfAddOns, actualNumberOfAddOns))
	}
}
