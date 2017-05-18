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
	"testing"
)

func Test_unix_oc_path(t *testing.T) {
	shellConfig, err := getOcShellConfig("/Users/john/.minishift/cache/oc/v1.5.0/oc", "")
	if err != nil {
		t.Fatalf("Unexepcted error: %s", err)
	}

	if shellConfig.OcDirPath != "/Users/john/.minishift/cache/oc/v1.5.0" {
		t.Fatalf("Unexepcted oc path: %s", shellConfig.OcDirPath)
	}
}
