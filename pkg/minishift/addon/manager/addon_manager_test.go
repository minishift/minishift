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
	"testing"

	"errors"
	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/cmd/testing/cli"
	"github.com/minishift/minishift/pkg/minishift/addon"
	"github.com/minishift/minishift/pkg/minishift/addon/command"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/stretchr/testify/assert"
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
	assert.Error(t, err, fmt.Sprintf("Creating the manager in directory '%s' should have failed", path))
	assert.Regexp(t, "^Unable to create addon manager", err.Error(), "Unexpected error message '%s'", err)
}

func Test_create_addon_manager(t *testing.T) {
	path := filepath.Join(basepath, "..", "..", "..", "..", "addons")

	manager, err := NewAddOnManager(path, make(map[string]*addon.AddOnConfig))
	assert.NoError(t, err, "Unexpected error creating manager in directory '%s'", path)

	addOns := manager.List()

	expectedNumberOfAddOns := 4
	assert.Len(t, addOns, expectedNumberOfAddOns)
}

func Test_installing_addon_for_non_existing_directory_returns_an_error(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-addon-manager-")
	defer os.RemoveAll(testDir)

	manager, err := NewAddOnManager(testDir, make(map[string]*addon.AddOnConfig))
	_, err = manager.Install("foo", false)

	assert.Error(t, err, "Creation of addon should have failed")
	assert.Regexp(t, "^The source of a addon needs to be a directory", err.Error(), "Unexpected error message")
}

func Test_invalid_addons_get_skipped(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-addon-manager-")
	assert.NoError(t, err, "Error creating temp directory")
	defer os.RemoveAll(testDir)

	// create a valid addon
	addOn1Path := filepath.Join(testDir, "addon1")
	err = os.Mkdir(addOn1Path, 0777)
	assert.NoError(t, err, "Error in creating directory for addon")

	err = ioutil.WriteFile(filepath.Join(addOn1Path, "anyuid.addon"), []byte(anyuid), 0777)
	assert.NoError(t, err, "Error in writing to addon file")

	// create a invalid addon
	addOn2Path := filepath.Join(testDir, "addon2")
	err = os.Mkdir(addOn2Path, 0777)
	assert.NoError(t, err, "Error in creating directory for addon")
	_, err = os.Create(filepath.Join(addOn2Path, "foo.addon"))
	assert.NoError(t, err, "Error in creating file for addon")

	manager, err := NewAddOnManager(testDir, make(map[string]*addon.AddOnConfig))

	assert.NoError(t, err, "Error in getting new addon manager")

	addOns := manager.List()

	expectedNumberOfAddOns := 1
	assert.Len(t, addOns, expectedNumberOfAddOns)
}

func TestAddVarDefaultsToContext(t *testing.T) {
	context, _ := command.NewExecutionContext(nil, nil)
	expectedVarName := "FOO"
	expectedVarDefaultValue := "foo"
	varDefault := fmt.Sprintf("%s=%s", expectedVarName, expectedVarDefaultValue)

	testAddonMap := getTestAddonMap("test", "test description", expectedVarName, varDefault, "")

	addOnMeta := getAddOnMetadata(testAddonMap, t)
	addOn := addon.NewAddOn(addOnMeta, []command.Command{}, []command.Command{}, "")

	addVarDefaultsToContext(addOn, context)
	assert.EqualValues(t, []string{expectedVarName}, context.Vars())
}

func TestVerifyValidRequiredVariablesInContext(t *testing.T) {
	context, _ := command.NewExecutionContext(nil, nil)
	expectedVarName := "FOO"
	expectedVarValue := "foo"

	testAddonMap := getTestAddonMap("test", "test description", expectedVarName, "", "")

	addOnMeta := getAddOnMetadata(testAddonMap, t)
	addOn := addon.NewAddOn(addOnMeta, []command.Command{}, []command.Command{}, "")

	// Add variable name to context
	context.AddToContext(expectedVarName, expectedVarValue)
	err := verifyRequiredVariablesInContext(context, addOn.MetaData())
	assert.NoError(t, err)
}

func TestVerifyMissingRequiredVariablesInContext(t *testing.T) {
	context, _ := command.NewExecutionContext(nil, nil)
	expectedVarName := "FOO"
	expectedErrMsg := "The variable(s) 'FOO' are required by the add-on, but are not defined in the context"

	testAddonMap := getTestAddonMap("test", "test description", expectedVarName, "", "")

	addOnMeta := getAddOnMetadata(testAddonMap, t)
	addOn := addon.NewAddOn(addOnMeta, []command.Command{}, []command.Command{}, "")

	err := verifyRequiredVariablesInContext(context, addOn.MetaData())
	assert.EqualError(t, err, expectedErrMsg)
}

type FakeSSHDockerCommander struct {
	docker.DockerCommander
	provision.SSHCommander
}

func (f *FakeSSHDockerCommander) SSHCommand(args string) (string, error) {
	return "openshift v3.6.1+008f2d5\nkubernetes v1.6.1+5115d708d7\netcd 3.2.1", nil
}

var expectedApplyAddonOutput = `-- Applying addon 'testaddon':
This testaddon is having variable TEST with foo value
`

func TestApplyAddon(t *testing.T) {
	path := filepath.Join(basepath, "..", "..", "..", "..", "test", "testdata", "testaddons")
	manager, err := NewAddOnManager(path, make(map[string]*addon.AddOnConfig))
	assert.NoError(t, err, "Unexpected error creating manager in directory '%s'", path)

	testaddon := manager.Get("testaddon")
	context, _ := command.NewExecutionContext(nil, &FakeSSHDockerCommander{})

	tee := cli.CreateTee(t, false)
	manager.ApplyAddOn(testaddon, context)
	tee.Close()

	assert.Equal(t, expectedApplyAddonOutput, tee.StdoutBuffer.String())
}

var expectedRemoveAddonOutput = `-- Removing addon 'testaddon':
Removing testaddon with variable TEST of foo value
`

func TestRemoveAddon(t *testing.T) {
	path := filepath.Join(basepath, "..", "..", "..", "..", "test", "testdata", "testaddons")
	manager, err := NewAddOnManager(path, make(map[string]*addon.AddOnConfig))
	assert.NoError(t, err, "Unexpected error creating manager in directory '%s'", path)

	testaddon := manager.Get("testaddon")
	context, _ := command.NewExecutionContext(nil, &FakeSSHDockerCommander{})

	tee := cli.CreateTee(t, false)
	manager.RemoveAddOn(testaddon, context)
	tee.Close()

	assert.Equal(t, expectedRemoveAddonOutput, tee.StdoutBuffer.String())
}

var expectedInvalidAddonOperationError = errors.New("The variable(s) 'TEST' are required by the add-on, but are not defined in the context")

func TestApplyInvalidAddon(t *testing.T) {
	path := filepath.Join(basepath, "..", "..", "..", "..", "test", "testdata", "testaddons")
	manager, err := NewAddOnManager(path, make(map[string]*addon.AddOnConfig))
	assert.NoError(t, err, "Unexpected error creating manager in directory '%s'", path)

	testaddon := manager.Get("invalidaddon")
	context, err := command.NewExecutionContext(nil, &FakeSSHDockerCommander{})
	assert.NoError(t, err, "Unexpected error creating new execution context")

	err = manager.ApplyAddOn(testaddon, context)
	assert.EqualError(t, expectedInvalidAddonOperationError, err.Error())
}

func TestRemoveInvalidAddon(t *testing.T) {
	path := filepath.Join(basepath, "..", "..", "..", "..", "test", "testdata", "testaddons")
	manager, err := NewAddOnManager(path, make(map[string]*addon.AddOnConfig))
	assert.NoError(t, err, "Unexpected error creating manager in directory '%s'", path)

	testaddon := manager.Get("invalidaddon")
	context, err := command.NewExecutionContext(nil, &FakeSSHDockerCommander{})
	assert.NoError(t, err, "Unexpected error creating new execution context")

	err = manager.RemoveAddOn(testaddon, context)
	assert.EqualError(t, expectedInvalidAddonOperationError, err.Error())
}

func getTestAddonMap(name, description, requireVar, varDefault, openshiftVersion string) map[string]interface{} {
	testAddonMap := make(map[string]interface{})
	testAddonMap["Name"] = name
	testAddonMap["Description"] = []string{description}
	if requireVar != "" {
		testAddonMap["Required-Vars"] = fmt.Sprintf("%s", requireVar)
	}
	if varDefault != "" {
		testAddonMap["Var-Defaults"] = fmt.Sprintf("%s", varDefault)
	}
	if openshiftVersion != "" {
		testAddonMap["OpenShift-Version"] = openshiftVersion
	}
	return testAddonMap
}

func getAddOnMetadata(testMap map[string]interface{}, t *testing.T) addon.AddOnMeta {
	addOnMeta, err := addon.NewAddOnMeta(testMap)
	if err != nil {
		t.Fatal(fmt.Sprintf("No error expected, but got: \"%s\"", err.Error()))
	}

	return addOnMeta
}
