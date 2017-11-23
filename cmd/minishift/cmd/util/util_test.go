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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidProfileName(t *testing.T) {
	testData := []struct {
		profileName string
		expected    bool
	}{{"test", true},
		{".", false},
		{"@foo", false},
		{"foo123", true},
		{"123foo", true},
		{"foo_123", true},
		{"foo@123", false},
		{"", false},
		{"a", true},
		{"a-", true},
		{"a_", true},
		{"1", true},
		{"_", false},
		{"-", false},
		{"-foo", false},
		{"--hell", false},
		{"foo-123", true}}
	for _, v := range testData {
		got := IsValidProfileName(v.profileName)
		assert.Equal(t, v.expected, got)
	}
}
