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
	"github.com/minishift/minishift/pkg/minishift/addon/command"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
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

func Test_successful_parsing_of_addon_dir(t *testing.T) {
	path := filepath.Join(basepath, "..", "..", "..", "..", "addons", "anyuid")
	addOn, err := testParser.Parse(path)

	if err != nil {
		t.Fatal("Unexpected error parsing addon content: " + err.Error())
	}

	expectedName := "anyuid"
	if addOn.MetaData().Name() != expectedName {
		t.Fatalf("Unexpected addon name: '%s'. Expected '%s'", addOn.MetaData().Name(), expectedName)
	}

	expectedNumberOfCommands := 5
	if len(addOn.Commands()) != expectedNumberOfCommands {
		t.Errorf("Unexpected number of commands. Found %d, but expected %d", len(addOn.Commands()), expectedNumberOfCommands)
	}

	_, ok := addOn.Commands()[0].(*command.OcCommand)
	if !ok {
		t.Fatalf("Expected an OcCommand. Got %s", reflect.TypeOf(addOn.Commands()[0]).Name())
	}
}

func Test_empty_lines_and_comments_in_content_are_ignored(t *testing.T) {
	_, commands, err := testParser.parseAddOnContent(strings.NewReader(addOnWithEmptyLinesAndCommments))

	if err != nil {
		t.Fatal("Unexpected error parsing addon content: " + err.Error())
	}

	expectedNumberOfCommands := 3
	if len(commands) != expectedNumberOfCommands {
		t.Errorf("Unexpected number of commands. Found %d, but expected %d", len(commands), expectedNumberOfCommands)
	}
}

func Test_addon_with_unknown_command_returns_error(t *testing.T) {
	_, _, err := testParser.parseAddOnContent(strings.NewReader(addOnWithUnknownCommand))

	if err == nil {
		t.Fatal("Expected error, got none")
	}

	expectedError := "Unable to process command: 'snafu'"
	if err.Error() != expectedError {
		t.Fatalf("Expected '%s'. Got '%s'", expectedError, err.Error())
	}
}

func Test_non_existend_addon_directory_creates_error(t *testing.T) {
	_, err := testParser.Parse("foo")

	if err == nil {
		t.Fatal("The parsing should have returned an error")
	}

	_, ok := err.(ParseError)
	if !ok {
		t.Fatalf("Expected an ParseError. Got %s", reflect.TypeOf(err))
	}
}

func Test_successful_parsing_of_content(t *testing.T) {
	meta, commands, err := testParser.parseAddOnContent(strings.NewReader(anyuid))

	if err != nil {
		t.Fatal("Unexpected error parsing addon content: " + err.Error())
	}

	expectedName := "anyuid"
	if meta.Name() != expectedName {
		t.Fatal(fmt.Sprintf("Unexpected addon name: '%s'. Expected: '%s'", meta.Name(), expectedName))
	}

	expectedDescription := "Allows authenticated users to run images to run with USER as per Dockerfile"
	if meta.Name() != expectedName {
		t.Fatal(fmt.Sprintf("Unexpected addon description: '%s'. Expected: '%s'", meta.Description(), expectedDescription))
	}

	if len(commands) != 1 {
		t.Errorf("Unexpected number of commands. Found %d, but expected 1", len(commands))
	}

	_, ok := commands[0].(*command.OcCommand)
	if !ok {
		t.Fatalf("Expected an OcCommand. Got %s", reflect.TypeOf(commands[0]))
	}
}

func Test_all_command_types_recognized(t *testing.T) {
	_, commands, _ := testParser.parseAddOnContent(strings.NewReader(oneOfAll))

	expectedNumberOfCommands := 4
	if len(commands) != expectedNumberOfCommands {
		t.Errorf("Unexpected number of commands. Found %d, but expected %d", len(commands), expectedNumberOfCommands)
	}

	_, ok := commands[0].(*command.OcCommand)
	if !ok {
		t.Fatalf("Expected an OcCommand. Got %s", reflect.TypeOf(commands[0]))
	}

	_, ok = commands[1].(*command.OpenShiftCommand)
	if !ok {
		t.Fatalf("Expected an OpenShiftCommand. Got %s", reflect.TypeOf(commands[1]))
	}

	_, ok = commands[2].(*command.SleepCommand)
	if !ok {
		t.Fatalf("Expected an SleepCommand. Got %s", reflect.TypeOf(commands[2]))
	}

	_, ok = commands[3].(*command.DockerCommand)
	if !ok {
		t.Fatalf("Expected an DockerCommand. Got %s", reflect.TypeOf(commands[3]))
	}
}

func Test_metadata_creation_requires_name_attribute(t *testing.T) {
	_, _, err := testParser.parseAddOnContent(strings.NewReader(noName))

	if err == nil {
		t.Fatal("The parsing should have returned an error")
	}

	tag := "Name"
	if !strings.Contains(err.Error(), tag) {
		t.Fatal(fmt.Sprintf("The error message should contain the tag '%s', but was '%s'", tag, err.Error()))
	}
}

func Test_metadata_creation_requires_description_attribute(t *testing.T) {
	_, _, err := testParser.parseAddOnContent(strings.NewReader(noDescription))

	if err == nil {
		t.Fatal("The parsing should have returned an error")
	}

	tag := "Description"
	if !strings.Contains(err.Error(), tag) {
		t.Fatal(fmt.Sprintf("The error message should contain the tag '%s', but was '%s'", tag, err.Error()))
	}
}

func Test_at_least_one_addon_file_required(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-addon-parser-")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(testDir)

	_, err = os.Create(filepath.Join(testDir, "test.json"))
	if err != nil {
		t.Error(err)
	}

	addon, err := testParser.Parse(testDir)

	if err == nil {
		t.Fatal("Expected a parse error for directory w/o addon file")
	}

	expectedError := fmt.Sprintf(noAddOnDefinitionFoundError, testDir)
	if err.Error() != expectedError {
		t.Fatal(fmt.Sprintf("Unexpected error message. Got '%s'. Expected '%s'", err.Error(), expectedError))
	}

	if addon != nil {
		t.Fatal("No addon should have been returned")
	}
}

func Test_multiple_addon_definitions_create_error(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-addon-parser-")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(testDir)

	addonFiles := []string{"first.addon", "second.addon"}
	for _, f := range addonFiles {
		_, err = os.Create(filepath.Join(testDir, f))
		if err != nil {
			t.Error(err)
		}
	}

	addon, err := testParser.Parse(testDir)

	if err == nil {
		t.Fatal("Expected a parse error for directory w/o addon file")
	}

	expectedError := fmt.Sprintf(multipleAddOnDefinitionsError, strings.Join(addonFiles, ", "))
	if err.Error() != expectedError {
		t.Fatal(fmt.Sprintf("Unexpected error message. Got '%s'. Expected '%s'", err.Error(), expectedError))
	}

	if addon != nil {
		t.Fatal("No addon should have been returned")
	}
}
