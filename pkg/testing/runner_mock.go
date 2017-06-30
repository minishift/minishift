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

package testing

import (
	"io"
	"strings"
	"testing"
)

var (
	expectCommands []ExpectedCommand
	t              *testing.T
)

type Runner struct{}

type FakeRunner struct {
	*Runner
}

type ExpectedCommand struct {
	commandArgs     string
	commandToOutput string
}

func NewFakeRunner(test *testing.T) *FakeRunner {
	fakeRunner := &FakeRunner{}
	t = test
	return fakeRunner
}

func (f *FakeRunner) ExpectAndReturn(input string, output string) {
	expectCommands = append(expectCommands, ExpectedCommand{commandArgs: input, commandToOutput: output})
}

/*func (f *FakeRunner) Verify() {

}*/

func (f *Runner) Output(command string, args ...string) ([]byte, error) {
	argStringOption := strings.Join(args, " ")

	if expectCommands[0].commandArgs == argStringOption {
		commandToOutput := expectCommands[0].commandToOutput
		expectCommands = expectCommands[1:]
		return []byte(commandToOutput), nil
	}

	t.Fatalf("Expected: %+v, Got: %+v\n", argStringOption, expectCommands[0].commandArgs)
	return nil, nil
}

func (f *Runner) Run(stdOut io.Writer, stdErr io.Writer, commandPath string, args ...string) int {
	return 0
}
