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
	minishiftTesting "github.com/minishift/minishift/pkg/testing"
	"testing"
)

func Test_Metadata_creation_succeeds(t *testing.T) {
	testName := "foo"
	testDescription := []string{"bar"}

	testMap := make(map[string]interface{})
	testMap["Name"] = testName
	testMap["Description"] = testDescription

	addOnMeta := getAddOnMeta(testMap, t)

	if addOnMeta.Name() != testName {
		t.Fatal(fmt.Sprintf("Expected addon name '%s', but got '%s'", testName, addOnMeta.Name()))
	}

	minishiftTesting.AssertEqualSlice(addOnMeta.Description(), testDescription, t)
}

func Test_missing_name_returns_error(t *testing.T) {
	testDescription := []string{"bar"}

	testMap := make(map[string]interface{})
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

	testMap := make(map[string]interface{})
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
	testDescription := []string{"bar"}
	testUrl := "http://my.addon.io"

	testMap := make(map[string]interface{})
	testMap["Name"] = testName
	testMap["Description"] = testDescription
	testMap["Url"] = testUrl

	addOnMeta := getAddOnMeta(testMap, t)

	if addOnMeta.GetValue("Url") != testUrl {
		t.Fatal(fmt.Sprintf("Expected addon value '%s', but got '%s'", testName, addOnMeta.GetValue("Url")))
	}
}

func Test_required_vars_meta_is_extracted(t *testing.T) {
	testMap := make(map[string]interface{})
	testMap["Name"] = "acme"
	testMap["Description"] = []string{"Acme Add-on"}
	testMap["Required-Vars"] = " USER, access_token "

	addOnMeta := getAddOnMeta(testMap, t)

	minishiftTesting.AssertEqualSlice(addOnMeta.RequiredVars(), []string{"USER", "access_token"}, t)
}

func Test_required_vars_empty_if_not_specified(t *testing.T) {
	testMap := make(map[string]interface{})
	testMap["Name"] = "acme"
	testMap["Description"] = []string{"Acme Add-on"}

	addOnMeta := getAddOnMeta(testMap, t)

	minishiftTesting.AssertEqualSlice(addOnMeta.RequiredVars(), []string{}, t)
}

func getAddOnMeta(testMap map[string]interface{}, t *testing.T) AddOnMeta {
	addOnMeta, err := NewAddOnMeta(testMap)

	if err != nil {
		t.Fatal(fmt.Sprintf("No error expected, but got: '%s'", err.Error()))
	}

	return addOnMeta
}
