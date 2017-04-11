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

package addon

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/minishift/minishift/pkg/testing/cli"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

func Test_addon_name_must_be_specified_for_enable_command(t *testing.T) {
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)

	atexit.RegisterExitHandler(cli.CreateExitHandlerFunc(t, tee, 1, emptyEnableError))

	runEnableAddon(nil, nil)
}

func Test_unknown_name_for_enable_command_returns_error(t *testing.T) {
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	os.Mkdir(filepath.Join(tmpMinishiftHomeDir, "addons"), 0777)

	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)

	testAddOnName := "foo"
	expectedOut := fmt.Sprintf(noAddOnToEnableMessage+"\n", testAddOnName)
	atexit.RegisterExitHandler(cli.CreateExitHandlerFunc(t, tee, 0, expectedOut))

	runEnableAddon(nil, []string{testAddOnName})

	actualOut := tee.StdoutBuffer.String()
	if expectedOut != actualOut {
		t.Fatalf("Expected output '%s'. Got '%s'.", expectedOut, actualOut)
	}
}
