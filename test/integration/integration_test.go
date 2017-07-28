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
	"path/filepath"
	"regexp"
)

var (
	minishift       *Minishift
	minishiftArgs   string
	minishiftBinary string

	testDir string

	// Godog options
	godogFormat              string
	godogTags                string
	godogShowStepDefinitions bool
	godogStopOnFailure       bool
	godogNoColors            bool
	godogPaths               string
)

func TestMain(m *testing.M) {
	parseFlags()

	if godogTags != "" {
		godogTags += "&&"
	}
	runner := util.MinishiftRunner{CommandPath: minishiftBinary}
	if runner.IsCDK() {
		godogTags += "~minishift-only ~b2d-only"
	} else {
		godogTags += "~cdk-only ~rhel-only"
	}

	status := godog.RunWithOptions("minishift", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format:              godogFormat,
		Paths:               strings.Split(godogPaths, ","),
		Tags:                godogTags,
		ShowStepDefinitions: godogShowStepDefinitions,
		StopOnFailure:       godogStopOnFailure,
		NoColors:            godogNoColors,
	})

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func parseFlags() {
	flag.StringVar(&minishiftArgs, "minishift-args", "", "Arguments to pass to minishift")
	flag.StringVar(&minishiftBinary, "binary", "", "Path to minishift binary")

	flag.StringVar(&testDir, "test-dir", "", "Path to the directory in which to execute the tests")

	flag.StringVar(&godogFormat, "format", "progress", "Sets which format godog will use")
	flag.StringVar(&godogTags, "tags", "", "Tags for godog test")
	flag.BoolVar(&godogShowStepDefinitions, "definitions", false, "")
	flag.BoolVar(&godogStopOnFailure, "stop-on-failure ", false, "Stop when failure is found")
	flag.BoolVar(&godogNoColors, "no-colors", false, "Disable colors in godog output")
	flag.StringVar(&godogPaths, "paths", "./features", "")

	flag.Parse()
}

func FeatureContext(s *godog.Suite) {
	runner := util.MinishiftRunner{
		CommandArgs: minishiftArgs,
		CommandPath: minishiftBinary}

	minishift = &Minishift{runner: runner}

	// steps to execute `minishift` commands
	s.Step(`Minishift (?:has|should have) state "([^"]*)"$`,
		minishift.shouldHaveState)
	s.Step(`^executing "minishift ([^"]*)"$`,
		minishift.executingMinishiftCommand)
	s.Step(`^executing "minishift ([^"]*)" (succeeds|fails)$`,
		executingMinishiftCommandSucceedsOrFails)
	s.Step(`([^"]*) of command "minishift ([^"]*)" is equal to "([^"]*)"$`,
		commandReturnEquals)
	s.Step(`([^"]*) of command "minishift ([^"]*)" contains "([^"]*)"$`,
		commandReturnContains)

	// steps to execute `oc` commands
	s.Step(`^executing "oc ([^"]*)" retrying (\d+) times with wait period of (\d+) seconds$`,
		minishift.executingRetryingTimesWithWaitPeriodOfSeconds)
	s.Step(`^executing "oc ([^"]*)"$`,
		minishift.executingOcCommand)
	s.Step(`^executing "oc ([^"]*)" (succeeds|fails)$`,
		executingOcCommandSucceedsOrFails)

	// steps for scenario variables
	s.Step(`setting scenario variable "([^"]*)" to the stdout from executing "oc ([^"]*)"$`,
		minishift.setVariableExecutingOcCommand)
	s.Step(`^scenario variable "([^"]*)" should not be empty$`,
		variableShouldNotBeEmpty)

	// steps for rollout check
	s.Step(`^services "([^"]*)" rollout successfully$`,
		minishift.rolloutServicesSuccessfully)

	// steps to verify `stdout`, `stderr` and `exitcode` of commands executed
	s.Step(`([^"]*) should contain ([^"]*)$`,
		commandReturnShouldContain)
	s.Step(`([^"]*) should not contain "([^"]*)"$`,
		commandReturnShouldNotContain)
	s.Step(`([^"]*) should contain$`,
		commandReturnShouldContainContent)
	s.Step(`([^"]*) should not contain$`,
		commandReturnShouldNotContainContent)
	s.Step(`([^"]*) should equal "([^"]*)"$`,
		commandReturnShouldEqual)
	s.Step(`([^"]*) should equal$`,
		commandReturnShouldEqualContent)
	s.Step(`([^"]*) should be empty$`,
		commandReturnShouldBeEmpty)
	s.Step(`([^"]*) should not be empty$`,
		commandReturnShouldNotBeEmpty)
	s.Step(`([^"]*) should be valid ([^"]*)$`,
		shouldBeInValidFormat)

	// step for HTTP requests for minishift web console
	s.Step(`(body|status code) of HTTP request to "([^"]*)" (?:|at "([^"]*)" )(contains|is equal to) "([^"]*)"$`,
		verifyHTTPResponse)

	// step for HTTP requests for accessing application
	s.Step(`(body|status code) of HTTP request to "([^"]*)" of service "([^"]*)" in namespace "([^"]*)" (contains|is equal to) "([^"]*)"$`,
		getRoutingUrlAndVerifyHTTPResponse)

	// steps for verifying config file content
	s.Step(`^JSON config file "([^"]*)" (contains|does not contain) key "(.*)" with value matching "(.*)"$`,
		matchConfigValue)
	s.Step(`^JSON config file "([^"]*)" (has|does not have) key "(.*)"$`,
		checkConfigKey)

	s.BeforeSuite(func() {
		testDir = setUp()
		if runner.IsCDK() {
			runner.CDKSetup()
		}

		fmt.Println("Running Integration test in:", testDir)
		fmt.Println("Using binary:", minishiftBinary)
	})

	s.AfterSuite(func() {
		minishift.runner.EnsureDeleted()
	})
}

func setUp() string {
	if testDir == "" {
		testDir, _ = ioutil.TempDir("", "minishift-integration-test-")
	} else {
		ensureTestDirEmpty()
	}

	os.Setenv(constants.MiniShiftHomeEnv, testDir)
	return testDir
}

func ensureTestDirEmpty() {
	files, err := ioutil.ReadDir(testDir)
	if err != nil {
		fmt.Println(fmt.Sprintf("Unable to setup integration test directory: %v", err))
		os.Exit(1)
	}

	for _, file := range files {
		fullPath := filepath.Join(testDir, file.Name())
		if filepath.Base(file.Name()) == "cache" {
			fmt.Println(fmt.Sprintf("Keeping Minishift cache directory '%s' for test run.", fullPath))
			continue
		}
		os.RemoveAll(fullPath)
	}
}

//  To get values of nested keys, use following dot formating in Scenarios: key.nestedKey
//  If an array is expected, then expect: "[value1 value2 value3]"
//  If empty string, non existing value are expected, then expect "<nil>"
func getConfigValue(configPath string, keyPath string) (string, error) {
	var keyValue string
	data, err := ioutil.ReadFile(testDir + "/" + configPath)
	if err != nil {
		return "", fmt.Errorf("Cannot read config file: %v", err)
	}
	var values map[string]interface{}
	json.Unmarshal(data, &values)
	keyPathArray := strings.Split(keyPath, ".")
	for _, element := range keyPathArray {
		switch value := values[element].(type) {
		case map[string]interface{}:
			values = value
		case []interface{}, nil, string, float64:
			keyValue = fmt.Sprintf("%v", value)
		default:
			return "", errors.New("Unexpected type in JSON config, not supported by testsuite")
		}
	}
	return keyValue, nil
}

func matchConfigValue(configPath string, condition string, keyPath string, expectedValue string) error {
	expectedRegexp, err := regexp.Compile(expectedValue)
	if err != nil {
		return fmt.Errorf("Expected value must be a valid regular expression statement: ", err)
	}

	keyValue, err := getConfigValue(configPath, keyPath)
	if err != nil {
		return err
	}

	if (condition == "contains") && !expectedRegexp.MatchString(keyValue) {
		return fmt.Errorf("For key '%s' config contains unexpected value '%s'", keyPath, keyValue)
	} else if (condition == "does not contain") && expectedRegexp.MatchString(keyValue) {
		return fmt.Errorf("For key '%s' config contains value '%s', which it should not contain", keyPath, keyValue)
	}

	return nil
}

func checkConfigKey(configPath string, condition string, keyPath string) error {
	keyValue, err := getConfigValue(configPath, keyPath)
	if err != nil {
		return err
	}

	if (condition == "has") && (keyValue == "<nil>") {
		return fmt.Errorf("Config does not contain any value for key %s", keyPath)
	} else if (condition == "does not have") && (keyValue != "<nil>") {
		return fmt.Errorf("Config contains key %s with assigned value: %s", keyPath, keyValue)
	}

	return nil
}

func compareExpectedWithActualContains(expected string, actual string) error {
	if !strings.Contains(actual, expected) {
		return fmt.Errorf("Output did not match. Expected: %s, Actual: %s", expected, actual)
	}

	return nil
}

func compareExpectedWithActualNotContains(notexpected string, actual string) error {
	if strings.Contains(actual, notexpected) {
		return fmt.Errorf("Output did match. Not expected: %s, Actual: %s", notexpected, actual)
	}

	return nil
}

func compareExpectedWithActualEquals(expected string, actual string) error {
	if actual != expected {
		return fmt.Errorf("Output did not match. Expected: %s, Actual: %s", expected, actual)
	}

	return nil
}

func compareExpectedWithActualNotEquals(notexpected string, actual string) error {
	if actual == notexpected {
		return fmt.Errorf("Output did match. Not expected: %s, Actual: %s", notexpected, actual)
	}

	return nil
}

func getLastCommandOutput() CommandOutput {
	return commandOutputs[len(commandOutputs)-1]
}

func selectFieldFromLastOutput(commandField string) string {
	lastCommandOutput := getLastCommandOutput()
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

func commandReturnShouldNotContain(commandField string, notexpected string) error {
	return compareExpectedWithActualNotContains(notexpected, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldContainContent(commandField string, expected *gherkin.DocString) error {
	return compareExpectedWithActualContains(expected.Content, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldNotContainContent(commandField string, notexpected *gherkin.DocString) error {
	return compareExpectedWithActualNotContains(notexpected.Content, selectFieldFromLastOutput(commandField))
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

func commandReturnShouldNotBeEmpty(commandField string) error {
	return compareExpectedWithActualNotEquals("", selectFieldFromLastOutput(commandField))
}

func variableShouldNotBeEmpty(variableName string) error {
	return compareExpectedWithActualNotEquals("", minishift.GetVariableByName(variableName).Value)
}

type commandRunner func(string) error

func executingOcCommandSucceedsOrFails(command string, expectedResult string) error {
	return succeedsOrFails(minishift.executingOcCommand, command, expectedResult)
}

func executingMinishiftCommandSucceedsOrFails(command string, expectedResult string) error {
	return succeedsOrFails(minishift.executingMinishiftCommand, command, expectedResult)
}

func succeedsOrFails(execute commandRunner, command string, expectedResult string) error {
	err := execute(command)
	if err != nil {
		return err
	}

	lastCommandOutput := getLastCommandOutput()
	commandFailed := (lastCommandOutput.ExitCode != 0 ||
		len(lastCommandOutput.StdErr) != 0)

	if expectedResult == "succeeds" && commandFailed == true {
		return fmt.Errorf("Command '%s' did not execute successfully. cmdExit: %d, cmdErr: %s",
			lastCommandOutput.Command,
			lastCommandOutput.ExitCode,
			lastCommandOutput.StdErr)
	}
	if expectedResult == "fails" && commandFailed == false {
		return fmt.Errorf("Command executed successfully, however was expected to fail. cmdExit: %d, cmdErr: %s",
			lastCommandOutput.ExitCode,
			lastCommandOutput.StdErr)
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

func getRoutingUrlAndVerifyHTTPResponse(partOfResponse, urlRoot, serviceName, nameSpace, assertion, expected string) error {
	url := minishift.getRoute(serviceName, nameSpace)
	if urlRoot == "/" {
		return verifyHTTPResponse(partOfResponse, url, "", assertion, expected)
	} else if strings.HasPrefix(urlRoot, "/") {
		url := url + urlRoot
		return verifyHTTPResponse(partOfResponse, url, "", assertion, expected)
	} else {
		return fmt.Errorf("Wrong input format : %s. Input must start with /", urlRoot)
	}
	return nil
}
