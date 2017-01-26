/*
Copyright (C) 2016 Red Hat, Inc.

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
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestDisplayConsoleInMachineReadable(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-config-")
	if err != nil {
		t.Error()
	}
	defer os.RemoveAll(testDir)

	f, err := os.Create(testDir + "out.txt")
	if err != nil {
		t.Fatal("Error creating test file", err)
	}
	defer f.Close()

	os.Stdout = f

	expectedStdout := `HOST=192.168.1.1
PORT=8443
CONSOLE_URL=https://192.168.99.103:8443`

	displayConsoleInMachineReadable("192.168.1.1", "https://192.168.99.103:8443")
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal("Error setting offset back", err)
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal("Error reading file", err)
	}

	actualStdout := string(data)
	if strings.TrimSpace(actualStdout) != expectedStdout {
		t.Fatalf("Expected:\n '%s' \nGot\n '%s'", expectedStdout, actualStdout)
	}
}
