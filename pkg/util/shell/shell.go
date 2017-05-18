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
	"fmt"
	"github.com/docker/machine/libmachine/shell"
	"os"
)

type ShellConfig struct {
	Prefix    string
	Delimiter string
	Suffix    string
}

func GetShell(userShell string) (string, error) {
	if userShell != "" {
		return userShell, nil
	}
	return shell.Detect()
}

func FindNoProxyFromEnv() (string, string) {
	// first check for an existing lower case no_proxy var
	noProxyVar := "no_proxy"
	noProxyValue := os.Getenv(noProxyVar)

	// otherwise default to allcaps HTTP_PROXY
	if noProxyValue == "" {
		noProxyVar = "NO_PROXY"
		noProxyValue = os.Getenv("NO_PROXY")
	}
	return noProxyVar, noProxyValue
}

func GenerateUsageHint(userShell, cmdLine string) string {
	cmd := ""
	comment := "#"

	switch userShell {
	case "fish":
		cmd = fmt.Sprintf("eval (%s)", cmdLine)
	case "powershell":
		cmd = fmt.Sprintf("& %s | Invoke-Expression", cmdLine)
	case "cmd":
		cmd = fmt.Sprintf("\t@FOR /f \"tokens=*\" %%i IN ('%s') DO @call %%i", cmdLine)
		comment = "REM"
	case "emacs":
		cmd = fmt.Sprintf("(with-temp-buffer (shell-command \"%s\" (current-buffer)) (eval-buffer))", cmdLine)
		comment = ";;"
	default:
		cmd = fmt.Sprintf("eval $(%s)", cmdLine)
	}

	return fmt.Sprintf("%s Run this command to configure your shell: \n%s %s\n", comment, comment, cmd)
}

func GetPrefixSuffixDelimiterForSet(userShell string, pathVar bool) (prefix, suffix, delimiter string) {
	switch userShell {
	case "fish":
		prefix = "set -gx "
		suffix = "\";\n"
		delimiter = " \""
		if pathVar {
			delimiter = " $PATH "
		}
	case "powershell":
		prefix = "$Env:"
		suffix = "\"\n"
		if pathVar {
			suffix = ";" + prefix + "PATH" + suffix
		}
		delimiter = " = \""
	case "cmd":
		prefix = "SET "
		suffix = "\n"
		if pathVar {
			suffix = ";%PATH%" + suffix
		}
		delimiter = "="
	case "emacs":
		prefix = "(setenv \""
		suffix = "\")\n"
		delimiter = "\" \""
	default:
		prefix = "export "
		suffix = "\"\n"
		if pathVar {
			suffix = ":$PATH" + suffix
		}
		delimiter = "=\""
	}

	return
}

func GetPrefixSuffixDelimiterForUnSet(userShell string) (prefix, suffix, delimiter string) {
	switch userShell {
	case "fish":
		prefix = "set -e "
		suffix = ";\n"
		delimiter = ""
	case "powershell":
		prefix = `Remove-Item Env:\\`
		suffix = "\n"
		delimiter = ""
	case "cmd":
		prefix = "SET "
		suffix = "\n"
		delimiter = "="
	case "emacs":
		prefix = "(setenv \""
		suffix = ")\n"
		delimiter = "\" nil"
	default:
		prefix = "unset "
		suffix = "\n"
		delimiter = ""
	}

	return
}
