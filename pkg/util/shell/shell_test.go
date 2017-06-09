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

package shell

import (
	"os"
	"testing"
)

type shellTestCase struct {
	name         string
	cmdLine      string
	expectedHint string
}

type prefixSuffixDelimiterTestCase struct {
	shellName string
	prefix    string
	suffix    string
	pathVar   bool
	delimiter string
}

func TestUnknownShell(t *testing.T) {
	expectedShellName := "foo"
	defer func(shell string) { os.Setenv("SHELL", shell) }(os.Getenv("SHELL"))
	os.Setenv("SHELL", "/bin/"+expectedShellName)

	shell, err := GetShell("")
	if err != nil {
		t.Fatal("Unexpected err : ", err.Error())
	}

	if shell != expectedShellName {
		t.Fatalf("Expected shell to be %s but %s", expectedShellName, shell)
	}
}

func TestKnownShell(t *testing.T) {
	expectedShellName := "bash"
	shell, err := GetShell("bash")
	if err != nil {
		t.Fatal("Unexpected err : ", err.Error())
	}

	if shell != expectedShellName {
		t.Fatalf("Expected shell to be %s but %s", expectedShellName, shell)
	}
}

func TestNoProxyFromEnvWhenUsedLowerCase(t *testing.T) {
	expectedNoProxyVarName := "no_proxy"
	expectedNoProxyVarValue := "FOO_BAR"
	defer func(val string) { os.Setenv(expectedNoProxyVarName, val) }(os.Getenv(expectedNoProxyVarName))
	os.Setenv(expectedNoProxyVarName, expectedNoProxyVarValue)

	_, noProxyValue := FindNoProxyFromEnv()
	if noProxyValue != expectedNoProxyVarValue {
		t.Fatalf("Expected no proxy var value as %s but got %s", expectedNoProxyVarValue, noProxyValue)
	}
}

func TestNoProxyFromEnvWhenUsedUpperCase(t *testing.T) {
	expectedNoProxyVarName := "NO_PROXY"
	expectedNoProxyVarValue := "FOO_BAR"
	defer func(val string) { os.Setenv(expectedNoProxyVarName, val) }(os.Getenv(expectedNoProxyVarName))
	os.Setenv(expectedNoProxyVarName, expectedNoProxyVarValue)

	_, noProxyValue := FindNoProxyFromEnv()
	if noProxyValue != expectedNoProxyVarValue {
		t.Fatalf("Expected no proxy var value as %s but got %s", expectedNoProxyVarValue, noProxyValue)
	}
}

func TestGenerateUsageHint(t *testing.T) {
	testData := []shellTestCase{
		{
			name: "bash", cmdLine: "foo", expectedHint: `# Run this command to configure your shell:
# eval $(foo)
`,
		},
		{
			name: "fish", cmdLine: "foo", expectedHint: `# Run this command to configure your shell:
# eval (foo)
`,
		},
		{
			name: "powershell", cmdLine: "foo", expectedHint: `# Run this command to configure your shell:
# & foo | Invoke-Expression
`,
		},
		{
			name: "cmd", cmdLine: "foo", expectedHint: `REM Run this command to configure your shell:
REM 	@FOR /f "tokens=*" %i IN ('foo') DO @call %i
`,
		},
		{
			name: "emacs", cmdLine: "foo", expectedHint: `;; Run this command to configure your shell:
;; (with-temp-buffer (shell-command "foo" (current-buffer)) (eval-buffer))
`,
		},
	}

	for _, tt := range testData {
		hint := GenerateUsageHint(tt.name, tt.cmdLine)
		if tt.expectedHint != hint {
			t.Fatalf("Expected hint to be \n%s but got \n%s", tt.expectedHint, hint)
		}

	}
}
