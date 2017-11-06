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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/spf13/viper"
)

func TestIsoUrl(t *testing.T) {
	var dummyIsoFile = createDummyIsoFile()

	var isoURLCheck = []struct {
		in  string
		out bool
	}{
		{"http://github.com/minishift/minishift-centos/minishift-centos.iso", true},
		{"https://github.com/minishift//minishift/minishift-b2d.iso", true},
		{"blahblah/http://b2d.iso", false},
		{"blabityblah", false},
		{strings.Replace("file://"+dummyIsoFile, "\\", "/", -1), true},
		{"file://home/bar/baz.iso", false},
		{"file://minishift/hoxier/foo", false},
		{"/home/joey/chandler/iso.iso", false},
		{"file:///home/homie/foo", false},
		{"ram/rahim/anthony.iso", false},
		{"gopher/minishift/b2d", false},
		{"file://d:/gopher/minishift/b2d", false},
		{"c:/gohome/foo.iso", false},
		{"http:/github.com/minishift/minishift.iso", false},
		{"file:/home/gopher/minishift-b2d.iso", false},
	}

	for _, urlTest := range isoURLCheck {
		viper.Set(configCmd.ISOUrl.Name, urlTest.in)
		defer viper.Reset()
		ret := checkIsoURL()
		if ret != urlTest.out {
			t.Errorf("Expected '%t' for given url '%s'. Got '%t'.", urlTest.out, urlTest.in, ret)
		}
	}

	defer removeDummyIsoFile(dummyIsoFile)
}

func createDummyIsoFile() string {
	dir, err := ioutil.TempDir("", "minishift-test-")
	if err != nil {
		fmt.Errorf("Failed to create directory")
	}
	f, err := os.Create(filepath.Join(dir, "foo.iso"))
	defer f.Close()
	if err != nil {
		fmt.Errorf("Failed to create dummy ISO file")
	}
	return filepath.Join(dir, "foo.iso")
}

func removeDummyIsoFile(dummyIsoFile string) {
	err := os.RemoveAll(strings.TrimSuffix(dummyIsoFile, "foo.iso"))
	if err != nil {
		fmt.Errorf("Failed to remove temporary directory")
	}
}
