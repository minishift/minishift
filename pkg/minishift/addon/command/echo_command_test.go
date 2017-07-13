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

package command

import (
	"fmt"
	"testing"

	"github.com/minishift/minishift/pkg/testing/cli"
)

func Test_echo_command(t *testing.T) {
	testCases := []struct {
		echoCommand    string
		expectedOutput string
	}{
		{
			echoCommand:    "echo hello world",
			expectedOutput: "\nhello world",
		},
		{
			echoCommand:    "echo    good\n bye", // 3 extrace spaces are preserved
			expectedOutput: "\n   good\n bye",
		},
		{
			echoCommand:    "echo",
			expectedOutput: "\n",
		},
		{
			echoCommand:    "echo ",
			expectedOutput: "\n",
		},
		{
			echoCommand:    "echo  ", // single additional space
			expectedOutput: "\n ",
		},
	}

	for _, test := range testCases {
		tee, err := cli.NewTee(true)
		defer tee.Close()

		if err != nil {
			t.Fatal("Unexpected error: " + err.Error())
		}

		echo := NewEchoCommand(test.echoCommand)
		context := &FakeInterpolationContext{}
		echo.Execute(&ExecutionContext{interpolationContext: context})
		tee.Close()

		if tee.StdoutBuffer.String() != test.expectedOutput {
			t.Fatal(fmt.Sprintf("Unexpected output to stdout. Expected '%s', but got '%s'", test.expectedOutput, tee.StdoutBuffer.String()))
		}
	}
}

type FakeInterpolationContext struct {
}

func (ic *FakeInterpolationContext) AddToContext(key string, value string) error {
	return nil
}

func (ic *FakeInterpolationContext) RemoveFromContext(key string) error {
	return nil
}

func (ic *FakeInterpolationContext) Interpolate(cmd string) string {
	return cmd
}

func (ic *FakeInterpolationContext) Vars() []string {
	return []string{}
}
