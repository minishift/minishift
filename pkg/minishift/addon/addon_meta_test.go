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

	minishiftTesting "github.com/minishift/minishift/pkg/testing"
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

	expectedError := "Metadata does not contain a mandatory entry for 'Name'"
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

	expectedError := "Metadata does not contain a mandatory entry for 'Description'"
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
	requiredVars, _ := addOnMeta.RequiredVars()
	minishiftTesting.AssertEqualSlice(requiredVars, []string{"USER", "access_token"}, t)
}

func Test_required_vars_empty_if_not_specified(t *testing.T) {
	testMap := make(map[string]interface{})
	testMap["Name"] = "acme"
	testMap["Description"] = []string{"Acme Add-on"}

	addOnMeta := getAddOnMeta(testMap, t)
	requiredVars, _ := addOnMeta.RequiredVars()
	minishiftTesting.AssertEqualSlice(requiredVars, []string{}, t)
}

func Test_var_defaults_empty_if_not_specified(t *testing.T) {
	testMap := make(map[string]interface{})
	testMap["Name"] = "acme"
	testMap["Description"] = []string{"Acme Add-on"}

	addOnMeta := getAddOnMeta(testMap, t)
	varDefaults, _ := addOnMeta.VarDefaults()
	if len(varDefaults) != 0 {
		t.Fatal(fmt.Sprintf("Expected empty var default, but got: '%s'", len(varDefaults)))
	}
}

func Test_var_defaults_meta_is_extracted(t *testing.T) {
	testMap := make(map[string]interface{})
	testMap["Name"] = "acme"
	testMap["Description"] = []string{"Acme Add-on"}
	expectedDefaultEnv := "USER=foo"
	testMap["Var-Defaults"] = expectedDefaultEnv

	addOnMeta := getAddOnMeta(testMap, t)
	varDefaults, _ := addOnMeta.VarDefaults()
	if len(varDefaults) != 1 {
		t.Fatal(fmt.Sprintf("Expected one var default, but got: '%s'", len(varDefaults)))
	}
}

type varDefaultTestCase struct {
	expectedVarDefault string
	expectedError      string
}

func Test_invalid_var_defaults_meta(t *testing.T) {
	testCases := []varDefaultTestCase{
		{
			expectedVarDefault: "",
			expectedError:      "'' is not a well formed Var-Defaults definition.",
		},
		{
			expectedVarDefault: "USER=foo,,",
			expectedError:      "'USER=foo,,' is not a well formed Var-Defaults definition.",
		},
		{
			expectedVarDefault: "=ABC",
			expectedError:      "'=ABC' is not a well formed Var-Defaults definition.",
		},
		{
			expectedVarDefault: "USER=",
			expectedError:      "'USER=' is not a well formed Var-Defaults definition.",
		},
	}

	testMap := make(map[string]interface{})
	testMap["Name"] = "acme"
	testMap["Description"] = []string{"Acme Add-on"}

	for _, testCase := range testCases {
		testMap["Var-Defaults"] = testCase.expectedVarDefault
		_, err := NewAddOnMeta(testMap)
		if err == nil {
			t.Fatalf("Expected an error, got none for var default \"%s\"", testCase.expectedVarDefault)
		}

		if err.Error() != testCase.expectedError {
			t.Fatalf("Expected error \"%s\", but got \"%s\"", testCase.expectedError, err.Error())
		}
	}
}

func getAddOnMeta(testMap map[string]interface{}, t *testing.T) AddOnMeta {
	addOnMeta, err := NewAddOnMeta(testMap)

	if err != nil {
		t.Fatal(fmt.Sprintf("No error expected, but got: \"%s\"", err.Error()))
	}

	return addOnMeta
}

func Test_check_openshift_version_sematic(t *testing.T) {
	var versionTestData = []struct {
		OpenshiftVersion string
		expectedResult   bool
	}{
		{"3.6.0", true},
		{"3.4", true},
		{"<3.6.0", true},
		{"3.7.1.1", false},
		{"3.5.3=", false},
		{"3.5.30", true},
		{"3.10.3", true},
		{">=3.5.0, <=3.7.0", true},
		{">3.5, <=3.7.0", true},
		{"3.5.0, <=3.7.0", true},
		{">=3.5.0, <=3.7", true},
		{"v3", false},
		{"p3.6.0", false},
	}

	for _, versionTest := range versionTestData {
		if checkVersionSemantic(versionTest.OpenshiftVersion) != versionTest.expectedResult {
			t.Errorf("Expected: '%t', Got 'true' for version: %s", versionTest.expectedResult, versionTest.OpenshiftVersion)
		}
	}
}
