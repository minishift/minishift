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

package openshift

import (
	"testing"

	"github.com/minishift/minishift/pkg/util/os/atexit"

	"github.com/minishift/minishift/pkg/testing/cli"
)

func Test_unknown_patch_target_aborts_command(t *testing.T) {
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)

	atexit.RegisterExitHandler(cli.CreateExitHandlerFunc(t, tee, 1, unknownPatchTargetError))

	target = "foo"
	runPatch(nil, nil)
}

func Test_patch_cannot_be_empty(t *testing.T) {
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)

	atexit.RegisterExitHandler(cli.CreateExitHandlerFunc(t, tee, 1, emptyPatchError))

	target = "master"
	patch = ""
	runPatch(nil, nil)
}

func Test_patch_needs_to_be_valid_JSON(t *testing.T) {
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)

	atexit.RegisterExitHandler(cli.CreateExitHandlerFunc(t, tee, 1, invalidJSONError))

	target = "master"
	patch = "foo"
	runPatch(nil, nil)
}

func Test_patch_commands_needs_existing_vm(t *testing.T) {
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)

	atexit.RegisterExitHandler(cli.CreateExitHandlerFunc(t, tee, 1, nonExistentMachineError))

	target = "master"
	patch = "{\"corsAllowedOrigins\": \"*\"}"
	runPatch(nil, nil)
}
