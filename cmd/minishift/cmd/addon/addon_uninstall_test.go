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

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/minishift/minishift/cmd/testing/cli"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"

	"github.com/stretchr/testify/assert"
)

var adminUser string = `# Name: admin-user
# Description: Create admin user and assign the cluster-admin role to it.

oc create user admin --full-name=admin
oc adm policy add-cluster-role-to-user cluster-admin admin
`

func Test_addon_name_must_be_specified_for_uninstall_command(t *testing.T) {
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)

	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 1, emptyAddOnError))

	runUnInstallAddon(nil, nil)
}

func Test_uninstall_with_enable_flag_works(t *testing.T) {
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	os.Mkdir(filepath.Join(tmpMinishiftHomeDir, "addons"), 0777)

	// need to make sure config.json ends up in the tmp directory
	os.Mkdir(filepath.Join(tmpMinishiftHomeDir, "config"), 0777)
	constants.ConfigFile = filepath.Join(tmpMinishiftHomeDir, "config", "config.json")

	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)

	addOnManager := GetAddOnManager()

	assert.Empty(t, addOnManager.List())

	// create a dummy addon named as admin-user
	testAddOnDir := filepath.Join(tmpMinishiftHomeDir, "admin-user")
	os.Mkdir(testAddOnDir, 0777)
	err := ioutil.WriteFile(filepath.Join(testAddOnDir, "admin-user.addon"), []byte(adminUser), 0644)

	assert.NoError(t, err, "Unexpected error writing to file")

	enable = true
	runInstallAddon(nil, []string{testAddOnDir})

	runUnInstallAddon(nil, []string{"admin-user"})
	addOnManager = GetAddOnManager()

	assert.Empty(t, addOnManager.List())

}
