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
	"net"
	"strings"
	"testing"

	"github.com/minishift/minishift/test/integration/util"
)

func TestStartStop(t *testing.T) {
	runner := util.MinishiftRunner{
		Args:       *args,
		BinaryPath: *binaryPath,
		T:          t}
	defer runner.EnsureDeleted()

	runner.Start()
	runner.CheckStatus("Running")

	ip := runner.RunCommand("ip", true)
	ip = strings.TrimRight(ip, "\n")
	if net.ParseIP(ip) == nil {
		t.Fatalf("IP command returned an invalid address: %s", ip)
	}

	runner.RunCommand("stop", true)
	runner.CheckStatus("Stopped")

	runner.RunCommand("delete", true)
	runner.CheckStatus("Does Not Exist")
}
