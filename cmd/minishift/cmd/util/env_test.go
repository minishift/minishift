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

package util

import (
	"testing"
)

func TestReplaceEnv(t *testing.T) {
	env := []string{"HOME=/home/user/", "PATH=/bin:/sbin:/usr/bin",
		"LC_ALL=de_DE.UTF8", "MINISHIFTHOME=/home/user/.minishift"}
	replaced := ReplaceEnv(env, "LC_ALL", "C")

	if env[2] == "LC_ALL=de_DE.UTF8" && replaced[2] != "LC_ALL=C" {
		t.Fatalf("Environment variable did not get replaced: '%s', '%s'", env[2], replaced[2])
	}

	if len(env) != len(replaced) {
		t.Fatalf("Environment variables do not match length: '%s', '%s'", len(env), len(replaced))
	}
}
