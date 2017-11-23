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

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

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
		assert.Equal(t, testCase.expectedResult, actualResult)
	}
}

func TestRemove(t *testing.T) {
	var testCases = []struct {
		slice    []string
		element  string
		expected []string
	}{
		{[]string{"a", "b", "c"}, "b", []string{"a", "c"}},
		{[]string{"a", "b", "c"}, "", []string{"a", "b", "c"}},
		{[]string{"a", "b", "c"}, "d", []string{"a", "b", "c"}},
		{[]string{}, "a", []string{}},
	}

	for _, testCase := range testCases {
		actual := Remove(testCase.slice, testCase.element)
		assert.EqualValues(t, testCase.expected, actual)
	}
}

func TestHasMatcher(t *testing.T) {
	var testCases = []struct {
		testString     string
		matcher        func(string) bool
		expectedResult bool
	}{
		{"abc", HasLetters, true},
		{"123", HasNumbers, true},
		{"abc", HasNumbers, false},
		{"123", HasLetters, false},
		{"abc123", HasLetters, true},
		{"abc123", HasNumbers, true},
		{"abc", HasOnlyLetters, true},
		{"123", HasOnlyNumbers, true},
		{"abc123", HasOnlyLetters, false},
		{"abc123", HasOnlyNumbers, false},
		{"!@#$%^&*()", HasLetters, false},
		{"!@#$%^&*()", HasNumbers, false},
		{"!@#$%^&*()", HasOnlyLetters, false},
		{"!@#$%^&*()", HasOnlyNumbers, false},
		{"Hello, World!", HasOnlyLetters, false},
		{"255.255.255.0", HasOnlyNumbers, false},
	}

	for _, testCase := range testCases {
		actualResult := testCase.matcher(testCase.testString)
		assert.Equal(t, testCase.expectedResult, actualResult)
	}
}

func TestGetMatcher(t *testing.T) {
	var testCases = []struct {
		testString     string
		matcher        func(string) string
		expectedResult string
	}{
		{"abc", GetOnlyLetters, "abc"},
		{"123", GetOnlyNumbers, "123"},
		{"abc123", GetOnlyLetters, "abc"},
		{"abc123", GetOnlyNumbers, "123"},
		{"!@#$%^&*()", GetOnlyLetters, ""},
		{"!@#$%^&*()", GetOnlyNumbers, ""},
		{"123GB", GetSignedNumbers, "123"},
		{"-123GB", GetSignedNumbers, "-123"},
	}

	for _, testCase := range testCases {
		actualResult := testCase.matcher(testCase.testString)
		assert.Equal(t, testCase.expectedResult, actualResult)
	}
}

func TestEscapeSingleQuote(t *testing.T) {
	experimentstring := map[string]string{
		`foo'bar`:  `foo'"'"'bar`,
		`foo''bar`: `foo'"'"''"'"'bar`,
		`foobar`:   `foobar`,
	}
	for pass, expected := range experimentstring {
		assert.Equal(t, expected, EscapeSingleQuote(pass))
	}
}

func TestSplitAndTrim(t *testing.T) {
	var testCases = []struct {
		testString    string
		separator     string
		expectedSlice []string
	}{
		{"foo bar", " ", []string{"foo", "bar"}},
		{"foo=bar", "=", []string{"foo", "bar"}},
		{"foobar", "", []string{"f", "o"}},
		{"foo:bar-baz", "-", []string{"foo:bar", "baz"}},
	}

	for _, testCase := range testCases {
		tokens, _ := SplitAndTrim(testCase.testString, testCase.separator)
		actualSlice := []string{tokens[0], tokens[1]}
		assert.EqualValues(t, testCase.expectedSlice, actualSlice)
	}
}
