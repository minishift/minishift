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
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

func ReadInputFromStdin(fieldlabel string) string {
	var value string
	fmt.Printf("%s: ", fieldlabel)
	fmt.Scanln(&value)
	return value
}

func IsTtySupported() bool {
	// https://github.com/golang/go/issues/12978
	return terminal.IsTerminal(int(syscall.Stdin))
}

func ReadPasswordFromStdin(fieldlabel string) string {
	var value string
	fmt.Printf("%s: ", fieldlabel)
	pwinput, err := terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
		fmt.Println("[HIDDEN]")
		value = string(pwinput)
	}
	return value
}

func AskForConfirmation(message string) bool {
	userConfirmation := ReadInputFromStdin(message +
		" Do you want to continue [y/N]?")
	return strings.ToUpper(userConfirmation) == "Y"
}
