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

package addon

import (
	"testing"

	"fmt"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/testing/cli"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"io/ioutil"
	"os"
	"path/filepath"
)

var anyuid string = `# Name: anyuid
# Description: Allows authenticated users to run images to run with USER as per Dockerfile

oc adm policy add-scc-to-group anyuid system:authenticated
`

func Test_source_of_addon_must_be_specified(t *testing.T) {
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	os.Mkdir(filepath.Join(tmpMinishiftHomeDir, "addons"), 0777)

	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)

	atexit.RegisterExitHandler(cli.CreateExitHandlerFunc(t, tee, 1, unspecifiedSourceError))

	runInstallAddon(nil, []string{})
}

func Test_install_with_enable_flag_works(t *testing.T) {
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	os.Mkdir(filepath.Join(tmpMinishiftHomeDir, "addons"), 0777)

	// need to make sure config.json ends up in the tmp directory
	os.Mkdir(filepath.Join(tmpMinishiftHomeDir, "config"), 0777)
	constants.ConfigFile = filepath.Join(tmpMinishiftHomeDir, "config", "config.json")

	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)

	addOnManager := GetAddOnManager()
	if len(addOnManager.List()) != 0 {
		t.Fatal(fmt.Sprintf("There should be no add-ons installed. Got got : %v", addOnManager.List()))
	}

	// create a dummy addon
	testAddOnDir := filepath.Join(tmpMinishiftHomeDir, "anyuid")
	os.Mkdir(testAddOnDir, 0777)
	err := ioutil.WriteFile(filepath.Join(testAddOnDir, "anyuid.addon"), []byte(anyuid), 0644)
	if err != nil {
		t.Fatal(fmt.Sprintf("Unexpected error writing to file: %v", err))
	}

	enable = true
	runInstallAddon(nil, []string{testAddOnDir})

	addOnManager = GetAddOnManager()
	if len(addOnManager.List()) != 1 {
		t.Fatal(fmt.Sprintf("There should be only one add-on. Got got : %v", addOnManager.List()))
	}

	if addOnManager.Get("anyuid").MetaData().Name() != "anyuid" {
		t.Fatal("The anyuid addon should be installed.")
	}

	if addOnManager.Get("anyuid").IsEnabled() != true {
		t.Fatal("The anyuid addon should be installed AND enabled")
	}
}
