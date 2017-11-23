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

package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/minishift/minishift/pkg/minishift/addon/command"
	minishiftTesting "github.com/minishift/minishift/pkg/testing"
	"github.com/stretchr/testify/assert"
)

var anyuid string = `# Name: anyuid
# Description: Allows authenticated users to run images to run with USER as per Dockerfile

oc adm policy add-scc-to-group anyuid system:authenticated
`

var oneOfAll string = `# Name: one-of-all
# Description: One command of each type

oc foo
openshift bar
sleep 1
docker snafo
`

var noName string = `# Description: One command of each type
oc foo
`

var noDescription string = `# Name: One command of each type
oc foo
`

var addOnWithMultilineDescriptionWithTabAndSpace string = `# Name: foo
# Description: This is one line
#   This is second line
#		This is third line
# This is forth line

# First run oc command
oc foo
`

var addOnWithMultilineDescriptionWithComment string = `# Hello
# Name: foo
# first line
#   :second line
# Description: This is one line
#   This is:second line
#		This is third line

# First run oc command
oc foo
`

var addOnWithEmptyLinesAndCommments string = `# Name: test
# Description: test

# First run oc command
oc foo

# Then run openshift command
openshift bar

 # Last run sleep command
sleep 1
`

var addOnWithUnknownCommand string = `# Name: test
# Description: test

# First run oc command
oc foo

# Then run openshift command
snafu
`

var testParser = NewAddOnParser()

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func Test_successful_parsing_of_addon_dir_without_remove_addon_file(t *testing.T) {
	path := filepath.Join(basepath, "..", "..", "..", "..", "addons", "anyuid")
	addOn, err := testParser.Parse(path)

	assert.NoError(t, err, "Error in parsing addon content")

	expectedName := "anyuid"
	assert.Equal(t, expectedName, addOn.MetaData().Name())

	expectedNumberOfCommands := 5
	assert.Len(t, addOn.Commands(), expectedNumberOfCommands)

	assert.Empty(t, addOn.RemoveCommands())

	_, ok := addOn.Commands()[0].(*command.OcCommand)
	assert.True(t, ok)
}

func Test_successful_parsing_of_addon_dir_with_remove_addon_file(t *testing.T) {
	path := filepath.Join(basepath, "..", "..", "..", "..", "addons", "admin-user")
	addOn, err := testParser.Parse(path)

	assert.NoError(t, err, "Error in parsing addon content")

	expectedName := "admin-user"
	assert.Equal(t, expectedName, addOn.MetaData().Name())

	expectedNumberOfCommands := 2
	assert.Len(t, addOn.Commands(), expectedNumberOfCommands)

	expectedNumberOfRemoveCommands := 2
	assert.Len(t, addOn.RemoveCommands(), expectedNumberOfRemoveCommands)

	_, ok := addOn.Commands()[0].(*command.OcCommand)
	assert.True(t, ok)
}

func Test_empty_lines_and_comments_in_content_are_ignored(t *testing.T) {
	_, commands, err := testParser.parseAddOnContent(strings.NewReader(addOnWithEmptyLinesAndCommments))
	assert.NoError(t, err, "Error in parsing addon content")

	expectedNumberOfCommands := 3
	assert.Len(t, commands, expectedNumberOfCommands)
}

func Test_multiple_lines_in_description_with_tab_and_space(t *testing.T) {
	meta, commands, err := testParser.parseAddOnContent(strings.NewReader(addOnWithMultilineDescriptionWithTabAndSpace))
	assert.NoError(t, err, "Error in parsing addon content")

	expectedDescription := []string{"This is one line", "This is second line", "This is third line", "This is forth line"}
	minishiftTesting.AssertEqualSlice(expectedDescription, meta.Description(), t)

	expectedNumberOfCommands := 1
	assert.Len(t, commands, expectedNumberOfCommands)
}

func Test_multiple_lines_in_description_with_comment(t *testing.T) {
	meta, commands, err := testParser.parseAddOnContent(strings.NewReader(addOnWithMultilineDescriptionWithComment))
	assert.NoError(t, err, "Error in parsing addon content")

	expectedName := "foo"
	assert.Equal(t, expectedName, meta.Name())

	expectedDescription := []string{"This is one line", "This is:second line", "This is third line"}
	assert.Equal(t, expectedDescription, meta.Description())
	expectedNumberOfCommands := 1
	assert.Len(t, commands, expectedNumberOfCommands)
}

func Test_addon_with_unknown_command_returns_error(t *testing.T) {
	_, _, err := testParser.parseAddOnContent(strings.NewReader(addOnWithUnknownCommand))
	assert.Error(t, err, "Error in parsing addon content")

	expectedError := "Unable to process command: 'snafu'"
	assert.EqualError(t, err, expectedError)
}

func Test_non_existend_addon_directory_creates_error(t *testing.T) {
	_, err := testParser.Parse("foo")
	assert.Error(t, err, "Error in parsing addon content")

	_, ok := err.(ParseError)
	assert.True(t, ok)
}

func Test_successful_parsing_of_content(t *testing.T) {
	meta, commands, err := testParser.parseAddOnContent(strings.NewReader(anyuid))
	assert.NoError(t, err, "Error in parsing addon content")

	expectedName := "anyuid"
	assert.Equal(t, expectedName, meta.Name())

	var expectedDescription = []string{"Allows authenticated users to run images to run with USER as per Dockerfile"}
	assert.Equal(t, expectedDescription, meta.Description())

	assert.Len(t, commands, 1)

	_, ok := commands[0].(*command.OcCommand)
	assert.True(t, ok)
}

func Test_all_command_types_recognized(t *testing.T) {
	_, commands, _ := testParser.parseAddOnContent(strings.NewReader(oneOfAll))

	expectedNumberOfCommands := 4
	assert.Len(t, commands, expectedNumberOfCommands)

	_, ok := commands[0].(*command.OcCommand)
	assert.True(t, ok)

	_, ok = commands[1].(*command.OpenShiftCommand)
	assert.True(t, ok)

	_, ok = commands[2].(*command.SleepCommand)
	assert.True(t, ok)

	_, ok = commands[3].(*command.DockerCommand)
	assert.True(t, ok)
}

func Test_metadata_creation_requires_name_attribute(t *testing.T) {
	_, _, err := testParser.parseAddOnContent(strings.NewReader(noName))
	assert.Error(t, err, "Error in parsing addon content")

	tag := "Name"
	assert.Contains(t, err.Error(), tag)
}

func Test_metadata_creation_requires_description_attribute(t *testing.T) {
	_, _, err := testParser.parseAddOnContent(strings.NewReader(noDescription))
	assert.Error(t, err, "Error in parsing addon content")

	tag := "Description"
	assert.Contains(t, err.Error(), tag)
}

func Test_at_least_one_addon_file_required(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-addon-parser-")
	assert.NoError(t, err, "Error in creating temp directory")
	defer os.RemoveAll(testDir)

	_, err = os.Create(filepath.Join(testDir, "test.json"))
	assert.NoError(t, err, "Error in creating 'test.json'")

	addon, err := testParser.Parse(testDir)
	assert.Error(t, err)

	expectedError := fmt.Sprintf(noAddOnDefinitionFoundError, testDir)
	assert.EqualError(t, err, expectedError)

	assert.Nil(t, addon, "No addon should have been returned")
}

func Test_multiple_addon_definitions_create_error(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-addon-parser-")
	assert.NoError(t, err, "Error creating temp directory")
	defer os.RemoveAll(testDir)

	addonFiles := []string{"first.addon", "second.addon"}
	for _, f := range addonFiles {
		_, err = os.Create(filepath.Join(testDir, f))
		assert.NoError(t, err, "Error in creating addon files in addons dir")
	}

	addon, err := testParser.Parse(testDir)
	assert.Error(t, err)

	expectedError := fmt.Sprintf(multipleAddOnDefinitionsError, strings.Join(addonFiles, ", "))
	assert.EqualError(t, err, expectedError)

	assert.Nil(t, addon, "No addon should have been returned")
}
