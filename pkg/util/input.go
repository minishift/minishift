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
	"syscall"

	"github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/ssh/terminal"
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

// Retrive password from OS native credential provider/keychain
func GetPasswordKeyring(username string) (string, error) {
	password, err := keyring.Get(constants.BinaryName, username)
	if err == keyring.ErrNotFound {
		return password, err
	}
	return password, nil
}

// Store the registration password in OS native keychain
func SetPasswordKeyring(username, password string) error {
	err := keyring.Set(constants.BinaryName, username, password)
	if err != nil {
		return err
	}
	return nil
}
