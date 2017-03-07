// +build integration

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

package integration

import (
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"

	"github.com/minishift/minishift/test/integration/util"
)

var lastCommandOutput CommandOutput

type CommandOutput struct {
	Command  string
	StdOut   string
	StdErr   string
	ExitCode int
}

type Minishift struct {
	runner util.MinishiftRunner
}

func (m *Minishift) shouldHaveState(expected string) error {
	actual := m.runner.GetStatus()
	if !strings.Contains(actual, expected) {
		return fmt.Errorf("Minishift state did not match. Expected: %s, Actual: %s", expected, actual)
	}

	return nil
}

func (m *Minishift) shouldHaveAValidIPAddress() error {
	ip, _, _ := m.runner.RunCommand("ip")
	ip = strings.TrimRight(ip, "\n")
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("IP command returned an invalid address: %s", ip)
	}

	return nil
}

func (m *Minishift) executingCommand(command string) error {
	// TODO: there must be smarter way to destruct
	cmdOut, cmdErr, cmdExit := m.runner.RunCommand(command)
	lastCommandOutput = CommandOutput{
		command,
		cmdOut,
		cmdErr,
		cmdExit,
	}

	// Beware: you are responsible to verify the lastCommandOutput!
	return nil
}

func compareExpectedWithActualContains(expected string, actual string) error {

	if !strings.Contains(actual, expected) {
		return fmt.Errorf("Output did not match. Expected: %s, Actual: %s", expected, actual)
	}

	return nil
}

func compareExpectedWithActualEquals(expected string, actual string) error {

	if actual != expected {
		return fmt.Errorf("Output did not match. Expected: %s, Actual: %s", expected, actual)
	}

	return nil
}

func selectFieldFromLastOutput(commandField string) string {
	outputField := ""
	switch commandField {
	case "stdout":
		outputField = lastCommandOutput.StdOut
	case "stderr":
		outputField = lastCommandOutput.StdErr
	case "exitcode":
		outputField = strconv.Itoa(lastCommandOutput.ExitCode)
	default:
		fmt.Errorf("Incorrect field type specified for comparison: %s", commandField)
	}
	return outputField
}

func (m *Minishift) commandReturnShouldContain(commandField string, expected string) error {
	return compareExpectedWithActualContains(expected, selectFieldFromLastOutput(commandField))
}

func (m *Minishift) commandReturnShouldContainContent(commandField string, expected *gherkin.DocString) error {
	return compareExpectedWithActualContains(expected.Content, selectFieldFromLastOutput(commandField))
}

func (m *Minishift) commandReturnShouldEqual(commandField string, expected string) error {
	return compareExpectedWithActualEquals(expected, selectFieldFromLastOutput(commandField))
}

func (m *Minishift) commandReturnShouldEqualContent(commandField string, expected *gherkin.DocString) error {
	return compareExpectedWithActualEquals(expected.Content, selectFieldFromLastOutput(commandField))
}

func (m *Minishift) commandReturnShouldBeEmpty(commandField string) error {
	return compareExpectedWithActualEquals("", selectFieldFromLastOutput(commandField))
}

func FeatureContext(s *godog.Suite) {
	var givenArgs = flag.String("minishift-args", "", "Arguments to pass to minishift")
	var givenPath = flag.String("binary", fmt.Sprintf("../../out/%s-amd64/minishift", runtime.GOOS), "Path to minishift binary")

	runner := util.MinishiftRunner{
		CommandArgs: *givenArgs,
		CommandPath: *givenPath}

	m := &Minishift{runner: runner}

	s.Step(`Minishift should have state "([^"]*)"`, m.shouldHaveState)
	s.Step(`Minishift should have a valid IP address`, m.shouldHaveAValidIPAddress)
	s.Step(`executing "minishift ([^"]*)"`, m.executingCommand)
	s.Step(`([^"]*) should contain ([^"]*)`, m.commandReturnShouldContain)
	s.Step(`([^"]*) should contain`, m.commandReturnShouldContainContent)
	s.Step(`([^"]*) should equal ([^"]*)`, m.commandReturnShouldEqual)
	s.Step(`([^"]*) should equal`, m.commandReturnShouldEqualContent)
	s.Step(`([^"]*) should be empty`, m.commandReturnShouldBeEmpty)

	s.BeforeScenario(func(interface{}) {
		testDir := setUp()
		os.RemoveAll(testDir)
	})

	s.AfterScenario(func(interface{}, error) {
		m.runner.EnsureDeleted()
	})

}

func TestMain(m *testing.M) {
	status := godog.RunWithOptions("minishift", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format: "progress",
		Paths:  []string{"features"},
	})

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
