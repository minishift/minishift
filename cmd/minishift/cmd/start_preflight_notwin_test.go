// +build !windows

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

	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/spf13/viper"
)

func TestIsoUrl(t *testing.T) {
	defer viper.Reset()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	dummyIsoFile := filepath.Join(currentDir, "..", "..", "..", "test", "testdata", "dummy.iso")

	err = createDummyIsoFile(dummyIsoFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer removeDummyIsoFile(dummyIsoFile)

	var isoURLCheck = sharedIsoURLChecks
	isoURLCheck = append(isoURLCheck, testData{"file://" + dummyIsoFile, true})

	for _, urlTest := range isoURLCheck {
		viper.Set(configCmd.ISOUrl.Name, urlTest.in)
		ret := checkIsoURL()
		if ret != urlTest.out {
			t.Errorf("Expected '%t' for given url '%s'. Got '%t'.", urlTest.out, urlTest.in, ret)
		}
	}
}
