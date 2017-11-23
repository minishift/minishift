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

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"strings"

	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestIsoUrlOnWindows(t *testing.T) {
	defer viper.Reset()
	currentDir, err := os.Getwd()
	assert.NoError(t, err, "Error getting working directory")

	dummyIsoFile := filepath.Join(currentDir, "..", "..", "..", "test", "testdata", "dummy.iso")
	dummyIsoFileWithForwardSlash := strings.Replace(dummyIsoFile, "\\", "/", -1)

	var isoURLCheck = sharedIsoURLChecks
	// right path, but wrong path seperator
	isoURLCheck = append(isoURLCheck, testData{"file://" + dummyIsoFile, false})
	// right path
	isoURLCheck = append(isoURLCheck, testData{"file://" + dummyIsoFileWithForwardSlash, true})

	for _, urlTest := range isoURLCheck {
		viper.Set(configCmd.ISOUrl.Name, urlTest.in)
		actual := checkIsoURL()
		assert.Equal(t, urlTest.out, actual)
	}
}
