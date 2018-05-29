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
	"errors"
	"fmt"
	"github.com/docker/machine/libmachine/shell"
	"os"
	"runtime"
	"strings"
)

var (
	supportedShell = []string{"bash", "fish", "powershell", "cmd", "emacs", "tcsh", "zsh"}
)

type ShellConfig struct {
	Prefix     string
	Delimiter  string
	Suffix     string
	PathSuffix string
}

func GetShell(userShell string) (string, error) {
	if userShell != "" {
		if !isSupportedShell(userShell) {
			return "", errors.New(fmt.Sprintf("'%s' is not a supported shell.\nSupported shells are %s.", userShell, strings.Join(supportedShell, ", ")))
		}
		return userShell, nil
	}
	return shell.Detect()
}

func isSupportedShell(userShell string) bool {
	for _, shell := range supportedShell {
		if userShell == shell {
			return true
		}
	}
	return false
}

func FindNoProxyFromEnv() (string, string) {
	// first check for an existing lower case no_proxy var
	noProxyVar := "no_proxy"
	noProxyValue := os.Getenv(noProxyVar)

	// otherwise default to allcaps NO_PROXY
	if noProxyValue == "" {
		noProxyVar = "NO_PROXY"
		noProxyValue = os.Getenv("NO_PROXY")
	}

	// On Windows, we've got to cheat a little bit because there is no way to
	// get a specificially lowercase or uppercase env var. Furthermore, if
	// both no_proxy and NO_PROXY are set, os.Getenv() will always get
	// the value for NO_PROXY because it has higher sorting precedence.
	// Effectively, that means that an uppercase env var on Windows always has
	// precedence. In addition, it is common practice to write env vars in
	// Windows in uppercase. All in all, this makes it reasonable to always
	// uppercase the NO_PROXY var for Windows.
	if runtime.GOOS == "windows" {
		noProxyVar = strings.ToUpper(noProxyVar)
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

	return fmt.Sprintf("%s Run this command to configure your shell:\n%s %s\n", comment, comment, cmd)
}

func GetPrefixSuffixDelimiterForSet(userShell string) (prefix, delimiter, suffix, pathSuffix string) {
	switch userShell {
	case "fish":
		prefix = "set -gx "
		delimiter = " \""
		suffix = "\";\n"
		pathSuffix = "\" $PATH;\n"
	case "powershell":
		prefix = "$Env:"
		delimiter = " = \""
		suffix = "\"\n"
		pathSuffix = ";" + prefix + "PATH" + suffix
	case "cmd":
		prefix = "SET "
		delimiter = "="
		suffix = "\n"
		pathSuffix = ";%PATH%" + suffix
	case "emacs":
		prefix = "(setenv \""
		delimiter = "\" \""
		suffix = "\")\n"
	default:
		prefix = "export "
		delimiter = "=\""
		suffix = "\"\n"
		pathSuffix = ":$PATH" + suffix
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
