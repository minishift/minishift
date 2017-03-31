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

package addon

import (
	"fmt"
	"testing"
)

func Test_Metadata_creation_succeeds(t *testing.T) {
	testName := "foo"
	testDescription := "bar"

	testMap := make(map[string]string)
	testMap["Name"] = testName
	testMap["Description"] = testDescription

	addOnMeta, err := NewAddOnMeta(testMap)

	if err != nil {
		t.Fatal(fmt.Sprintf("No error expected, but got: '%s'", err.Error()))
	}

	if addOnMeta.Name() != testName {
		t.Fatal(fmt.Sprintf("Expected addon name '%s', but got '%s'", testName, addOnMeta.Name()))
	}

	if addOnMeta.Description() != testDescription {
		t.Fatal(fmt.Sprintf("Expected addon description '%s', but got '%s'", testName, addOnMeta.Description()))
	}
}

func Test_missing_name_returns_error(t *testing.T) {
	testDescription := "bar"

	testMap := make(map[string]string)
	testMap["Description"] = testDescription

	_, err := NewAddOnMeta(testMap)

	if err == nil {
		t.Fatal("Expected an error, got none")
	}

	expectedError := "Metadata does not contain an mandatory entry for 'Name'"
	if err.Error() != expectedError {
		t.Fatal(fmt.Sprintf("Expected error '%s', but got '%s'", expectedError, err.Error()))
	}
}

func Test_missing_description_returns_error(t *testing.T) {
	testName := "foo"

	testMap := make(map[string]string)
	testMap["Name"] = testName

	_, err := NewAddOnMeta(testMap)

	if err == nil {
		t.Fatal("Expected an error, got none")
	}

	expectedError := "Metadata does not contain an mandatory entry for 'Description'"
	if err.Error() != expectedError {
		t.Fatal(fmt.Sprintf("Expected error '%s', but got '%s'", expectedError, err.Error()))
	}
}

func Test_optional_metadata_is_retained(t *testing.T) {
	testName := "foo"
	testDescription := "bar"
	testUrl := "http://my.addon.io"

	testMap := make(map[string]string)
	testMap["Name"] = testName
	testMap["Description"] = testDescription
	testMap["Url"] = testUrl

	addOnMeta, err := NewAddOnMeta(testMap)

	if err != nil {
		t.Fatal(fmt.Sprintf("No error expected, but got: '%s'", err.Error()))
	}

	if addOnMeta.GetValue("Url") != testUrl {
		t.Fatal(fmt.Sprintf("Expected addon value '%s', but got '%s'", testName, addOnMeta.GetValue("Url")))
	}
}
