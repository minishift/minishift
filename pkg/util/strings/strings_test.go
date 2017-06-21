/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package strings

import "testing"

func TestContains(t *testing.T) {
	var testCases = []struct {
		slice          []string
		element        string
		expectedResult bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "e", false},
		{nil, "b", false},
	}

	for _, testCase := range testCases {
		actualResult := Contains(testCase.slice, testCase.element)
		if actualResult != testCase.expectedResult {
			t.Errorf("Unexpected result for Contains(%v, %s). Expected '%t', got '%t'", testCase.slice, testCase.element, testCase.expectedResult, actualResult)
		}
	}
}

func TestEscapeSingleQuote(t *testing.T) {
	experimentstring := map[string]string{
		`foo'bar`:  `foo'"'"'bar`,
		`foo''bar`: `foo'"'"''"'"'bar`,
		`foobar`:   `foobar`,
	}
	for pass, expected := range experimentstring {
		if EscapeSingleQuote(pass) != expected {
			t.Fatalf("Expected '%s' Got '%s'", expected, EscapeSingleQuote(pass))
		}
	}
}
