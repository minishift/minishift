// +build integration

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

package integration

import (
	"fmt"
	"strings"
	"testing"

	"github.com/minishift/minishift/test/integration/util"
)

func TestStartWithDockerEnv(t *testing.T) {
	runner := util.MinishiftRunner{
		Args:       *args,
		BinaryPath: *binaryPath,
		T:          t}
	defer runner.EnsureDeleted()

	runner.RunCommand(fmt.Sprintf("start %s --docker-env=FOO=BAR --docker-env=BAZ=BAT", runner.Args), true)
	runner.EnsureRunning()

	profileContents := runner.RunCommand("ssh cat /var/lib/boot2docker/profile", true)
	fmt.Println(profileContents)
	for _, envVar := range []string{"FOO=BAR", "BAZ=BAT"} {
		if !strings.Contains(profileContents, fmt.Sprintf("export \"%s\"", envVar)) {
			t.Fatalf("Env var %s missing from file: %s.", envVar, profileContents)
		}
	}
}
