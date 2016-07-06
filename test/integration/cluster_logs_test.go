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
	"strings"
	"testing"

	"github.com/jimmidyson/minishift/pkg/minikube/constants"
	"github.com/jimmidyson/minishift/test/integration/util"
)

func TestClusterLogs(t *testing.T) {
	minikubeRunner := util.MinikubeRunner{
		Args:       *args,
		BinaryPath: *binaryPath,
		T:          t}
	minikubeRunner.EnsureRunning()

	logsCmdOutput := minikubeRunner.RunCommand("logs", true)
	//check for # of lines or check for strings
	logFiles := []string{constants.RemoteOpenShiftErrPath, constants.RemoteOpenShiftOutPath}
	for _, logFile := range logFiles {
		if !strings.Contains(logsCmdOutput, logFile) {
			t.Fatalf("Error in logsCmdOutput, expected to find: %s. Output: %s", logFile, logsCmdOutput)
		}
	}
}
