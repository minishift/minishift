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
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

func ReadInputFromStdin(fieldlabel string) string {
	var value string
	print(fmt.Sprintf("%s: ", fieldlabel))
	fmt.Scanln(&value)
	return value
}

func ReadPasswordFromStdin(fieldlabel string) string {
	var value string
	print(fmt.Sprintf("%s: ", fieldlabel))
	pwinput, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err == nil {
		println("[HIDDEN]")
		value = string(pwinput)
	}
	return value
}
