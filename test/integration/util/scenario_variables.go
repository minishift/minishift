/*
Copyright (C) 2018 Red Hat, Inc.

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
)

var commandVariables []commandVariable

type commandVariable struct {
	Name  string
	Value string
}

func ProcessVariables(command string) string {
	for _, variable := range commandVariables {
		command = strings.Replace(command, fmt.Sprintf("$(%s)", variable.Name), variable.Value, -1)
	}

	return command
}

func SetVariable(name string, value string) {
	commandVariables = append(commandVariables,
		commandVariable{
			name,
			value,
		})
}

func GetVariableByName(name string) string {
	for i := range commandVariables {
		variable := commandVariables[i]
		if variable.Name == name {
			return variable.Value
		}
	}

	return ""
}
