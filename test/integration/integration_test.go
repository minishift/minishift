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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"

	"crypto/tls"
	"errors"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/test/integration/util"
	"net/http"
)

var givenArgs, givenPath, testDir string
var minishift *Minishift

func TestMain(m *testing.M) {
	flag.StringVar(&givenArgs, "minishift-args", "", "Arguments to pass to minishift")
	flag.StringVar(&givenPath, "binary", "", "Path to minishift binary")

	var godogFormat = flag.String("format", "progress", "Sets which format godog will use")
	var godogTags = flag.String("tags", "", "Tags for godog test")
	var godogShowStepDefinitions = flag.Bool("definitions", false, "")
	var godogStopOnFailure = flag.Bool("stop-on-failure ", false, "Stop when failure is found")
	var godogNoColors = flag.Bool("no-colors", false, "Disable colors in godog output")
	var godogPaths = flag.String("paths", "./features", "")
	flag.Parse()

	status := godog.RunWithOptions("minishift", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format:              *godogFormat,
		Paths:               strings.Split(*godogPaths, ","),
		Tags:                *godogTags,
		ShowStepDefinitions: *godogShowStepDefinitions,
		StopOnFailure:       *godogStopOnFailure,
		NoColors:            *godogNoColors,
	})

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func FeatureContext(s *godog.Suite) {
	runner := util.MinishiftRunner{
		CommandArgs: givenArgs,
		CommandPath: givenPath}

	minishift = &Minishift{runner: runner}

	// steps to execute minishift commands
	s.Step(`Minishift (?:has|should have) state "([^"]*)"`, minishift.shouldHaveState)
	s.Step(`executing "minishift ([^"]*)"`, minishift.executingMinishiftCommand)
	s.Step(`executing "minishift ([^"]*)" (.*)$`, executingMinishiftCommandSucceedsOrFails)
	s.Step(`([^"]*) of command "minishift ([^"]*)" is equal to "([^"]*)"`, commandReturnEquals)
	s.Step(`([^"]*) of command "minishift ([^"]*)" contains "([^"]*)"`, commandReturnContains)

	// steps for running oc
	s.Step(`executing "oc ([^"]*)" retrying (\d+) times with wait period of (\d+) seconds$`, minishift.executingRetryingTimesWithWaitPeriodOfSeconds)
	s.Step(`executing "oc ([^"]*)`, minishift.executingOcCommand)
	s.Step(`executing "oc ([^"]*)" succeeds$`, minishift.executingOcCommandSucceeds)

	// steps to verify stdout and stderr of commands executed
	s.Step(`([^"]*) should contain ([^"]*)`, commandReturnShouldContain)
	s.Step(`([^"]*) should contain`, commandReturnShouldContainContent)
	s.Step(`([^"]*) should equal ([^"]*)`, commandReturnShouldEqual)
	s.Step(`([^"]*) should equal`, commandReturnShouldEqualContent)
	s.Step(`([^"]*) should be empty`, commandReturnShouldBeEmpty)
	s.Step(`([^"]*) should be valid ([^"]*)`, shouldBeInValidFormat)

	// step for HTTP requests
	s.Step(`(body|status code) of HTTP request to "([^"]*)" (?:|at "([^"]*)" )(contains|is equal to) "([^"]*)"`, verifyHTTPResponse)

	// steps for verifying config file content
	s.Step(`JSON config file "([^"]*)" (contains|does not contain) key "(.*)" with value "(.*)"`, configContains)
	s.Step(`JSON config file "([^"]*)" (contains|does not contain) key "(.*)"(.*)`, configContains)

	s.BeforeSuite(func() {
		testDir = setUp()
		if runner.IsCDK() {
			runner.CDKSetup()
		}
		fmt.Println("Running Integration test in:", testDir)
		fmt.Println("using binary:", givenPath)
	})

	s.AfterSuite(func() {
		minishift.runner.EnsureDeleted()
	})
}

func setUp() string {
	testDir, _ := ioutil.TempDir("", "minishift-integration-test-")
	os.Setenv(constants.MiniShiftHomeEnv, testDir)
	return testDir
}

//  To get values of nested keys, use following dot formating in Scenarios: key.nestedKey
//  If an array is expected, use following formating: [value1 value2 value3].
func configContains(configPath, expectingResult, expectedKeyPath, expectedValue string) error {
	data, err := ioutil.ReadFile(testDir + "/" + configPath)
	if err != nil {
		return fmt.Errorf("Cannot read config file: %s", err)
	}
	var values map[string]interface{}
	json.Unmarshal(data, &values)
	actualValue := ""
	keyPath := strings.Split(expectedKeyPath, ".")
	for _, element := range keyPath {
		switch value := values[element].(type) {
		case map[string]interface{}:
			values = value
		case []interface{}, nil, string, float64:
			actualValue = fmt.Sprintf("%v", value)
		default:
			return errors.New("Unexpected type in JSON config, not supported by testsuite")
		}
	}
	if (expectingResult == "contains") && (actualValue != expectedValue) {
		return fmt.Errorf("For key '%s' config contains unexpected value '%s'", expectedKeyPath, actualValue)
	}

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

func shouldBeInValidFormat(commandField, format string) error {
	result := selectFieldFromLastOutput(commandField)
	result = strings.TrimRight(result, "\n")
	switch format {
	case "URL":
		_, err := url.ParseRequestURI(result)
		if err != nil {
			return fmt.Errorf("Command did not returned URL in valid format: %s", result)
		}
	case "IP":
		if net.ParseIP(result) == nil {
			return fmt.Errorf("%s of previous command is not a valid IP address: %s", commandField, result)
		}
	default:
		return fmt.Errorf("Format %s not implemented.", format)
	}
	return nil
}

func commandReturnEquals(commandField, command, expected string) error {
	minishift.executingMinishiftCommand(command)
	return compareExpectedWithActualEquals(expected+"\n", selectFieldFromLastOutput(commandField))
}

func commandReturnContains(commandField, command, expected string) error {
	minishift.executingMinishiftCommand(command)
	return compareExpectedWithActualContains(expected, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldContain(commandField string, expected string) error {
	return compareExpectedWithActualContains(expected, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldContainContent(commandField string, expected *gherkin.DocString) error {
	return compareExpectedWithActualContains(expected.Content, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldEqual(commandField string, expected string) error {
	return compareExpectedWithActualEquals(expected, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldEqualContent(commandField string, expected *gherkin.DocString) error {
	return compareExpectedWithActualEquals(expected.Content, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldBeEmpty(commandField string) error {
	return compareExpectedWithActualEquals("", selectFieldFromLastOutput(commandField))
}

func executingMinishiftCommandSucceedsOrFails(command, expectedResult string) error {
	err := minishift.executingMinishiftCommand(command)
	if err != nil {
		return err
	}
	success := (lastCommandOutput.ExitCode != 0 || len(lastCommandOutput.StdErr) != 0)
	switch expectedResult {
	case "succeeds":
		if success {
			return fmt.Errorf("Command did not execute successfully. cmdExit: %d, cmdErr: %s", lastCommandOutput.ExitCode, lastCommandOutput.StdErr)
		}
	case "fails":
		if success {
			return fmt.Errorf("Command executed successfully, however was expected to fail. cmdExit: %d, cmdErr: %s", lastCommandOutput.ExitCode, lastCommandOutput.StdErr)
		}
	default:
		return fmt.Errorf("Expected result: %s not recognized, please use: 'fails' or 'succeeds'", expectedResult)
	}
	return nil
}

func verifyHTTPResponse(partOfResponse, url, urlSuffix, assertion, expected string) error {
	switch url {
	case "OpenShift":
		url = minishift.getOpenShiftUrl() + urlSuffix
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}
	response, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("Server returned error on url: %s", url)
	}
	defer response.Body.Close()
	var result string
	switch partOfResponse {
	case "body":
		html, _ := ioutil.ReadAll(response.Body)
		result = string(html[:])
	case "status code":
		result = fmt.Sprintf("%d", response.StatusCode)
	default:
		return fmt.Errorf("%s not implemented", partOfResponse)
	}

	switch assertion {
	case "contains":
		if !strings.Contains(result, expected) {
			return fmt.Errorf("%s of reponse from %s does not contain expected string. Expected: %s, Actual: %s", partOfResponse, url, expected, result)
		}
	case "is equal to":
		if result != expected {
			return fmt.Errorf("%s of response from %s is not equal to expected string. Expected: %s, Actual: %s", partOfResponse, url, expected, result)
		}
	default:
		return fmt.Errorf("Assertion type: %s is not implemented", assertion)
	}
	return nil
}
