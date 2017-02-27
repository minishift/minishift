// +build integration

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

package integration

import (
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/test/integration/util"
	"os"
	"strings"
	"testing"
)

func TestClusterSSH(t *testing.T) {
	testDir := setUp(t)
	defer os.RemoveAll(testDir)
	defer os.Unsetenv(constants.MiniShiftHomeEnv)

	runner := util.MinishiftRunner{
		Args:       *args,
		BinaryPath: *binaryPath,
		T:          t,
	}
	defer runner.EnsureDeleted()

	runner.EnsureRunning()

	expectedStr := "hello"
	sshCmdOutput := runner.RunCommand("ssh echo "+expectedStr, true)
	if !strings.Contains(sshCmdOutput, expectedStr) {
		t.Fatalf("Expected output from ssh to be: %s. Output was: %s", expectedStr, sshCmdOutput)
	}
}
