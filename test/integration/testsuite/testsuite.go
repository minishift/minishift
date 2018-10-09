// +build integration

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

package testsuite

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"gopkg.in/yaml.v2"

	testProxy "github.com/minishift/minishift/test/integration/proxy"
	"github.com/minishift/minishift/test/integration/util"
)

var (
	MinishiftInstance *Minishift
	minishiftArgs     string
	minishiftBinary   string

	testDir         string
	testDefaultHome string
	testResultDir   string
	isoName         string

	runBeforeFeature string
	testWithShell    string
	copyOcFrom       string

	// Godog options
	GodogFormat              string
	GodogTags                string
	GodogShowStepDefinitions bool
	GodogStopOnFailure       bool
	GodogNoColors            bool
	GodogPaths               string
)

func FeatureContext(s *godog.Suite) {
	runner := util.MinishiftRunner{
		CommandArgs: minishiftArgs,
		CommandPath: minishiftBinary}

	MinishiftInstance = &Minishift{runner: runner}

	// Checking Minishift status and profiles
	s.Step(`^Minishift (?:has|should have) state "(Does Not Exist|Running|Stopped)"$`,
		MinishiftInstance.shouldHaveState)
	s.Step(`^profile (.*) (?:has|should have) state "(Does Not Exist|Running|Stopped)"$`,
		MinishiftInstance.profileShouldHaveState)
	s.Step(`profile (.*) (?:is the|should be the) active profile$`,
		MinishiftInstance.isTheActiveProfile)

	// Execution of `minishift` commands
	s.Step(`^executing "minishift (.*)"$`,
		MinishiftInstance.ExecutingMinishiftCommand)
	s.Step(`^executing "minishift (.*)" (succeeds|fails)$`,
		ExecutingMinishiftCommandSucceedsOrFails)
	s.Step(`^([^"]*) of command "minishift (.*)" (is equal|is not equal) to "(.*)"$`,
		commandReturnEquals)
	s.Step(`^([^"]*) of command "minishift (.*)" (contains|does not contain) "(.*)"$`,
		commandReturnContains)

	// Execution of `oc` commands
	// will use default version of oc binary at stored at .minishift/cache/oc
	s.Step(`^executing "oc (.*)" retrying (\d+) times with wait period of (\d+) seconds$`,
		MinishiftInstance.executingRetryingTimesWithWaitPeriodOfSeconds)
	s.Step(`^executing "oc (.*)"$`,
		MinishiftInstance.ExecutingOcCommand)
	s.Step(`^executing "oc (.*)" (succeeds|fails)$`,
		ExecutingOcCommandSucceedsOrFails)

	// Command output verification
	// steps to verify `stdout`, `stderr` and `exitcode` of last executed command,
	// supports simple string verification and also regular expressions.
	s.Step(`^(stdout|stderr|exitcode) should contain "(.*)"$`,
		commandReturnShouldContain)
	s.Step(`^(stdout|stderr|exitcode) should not contain "(.*)"$`,
		commandReturnShouldNotContain)
	s.Step(`^(stdout|stderr|exitcode) should contain$`,
		commandReturnShouldContainContent)
	s.Step(`^(stdout|stderr|exitcode) should not contain$`,
		commandReturnShouldNotContainContent)
	s.Step(`^(stdout|stderr|exitcode) should equal "(.*)"$`,
		commandReturnShouldEqual)
	s.Step(`^(stdout|stderr|exitcode) should equal$`,
		commandReturnShouldEqualContent)
	s.Step(`^(stdout|stderr|exitcode) should be empty$`,
		commandReturnShouldBeEmpty)
	s.Step(`^(stdout|stderr|exitcode) should not be empty$`,
		commandReturnShouldNotBeEmpty)
	s.Step(`^(stdout|stderr|exitcode) should be valid (.*)$`,
		shouldBeInValidFormat)
	s.Step(`^(stdout|stderr|exitcode) should match "(.*)"$`,
		commandReturnShouldMatchRegex)
	s.Step(`^(stdout|stderr|exitcode) should not match "(.*)"$`,
		commandReturnShouldNotMatchRegex)
	s.Step(`^(stdout|stderr|exitcode) should match$`,
		commandReturnShouldMatchRegexContent)
	s.Step(`^(stdout|stderr|exitcode) should not match$`,
		commandReturnShouldNotMatchRegexContent)

	// Executing commands in shells
	// for testing of integration with specific shells, for example, to test oc-env and docker-env
	// commands which uses `export` command (and its cmd and powershell alternatives)
	s.Step(`^user starts shell instance on host machine$`,
		startHostShellInstance)
	s.Step(`^user closes shell instance on host machine$`,
		util.CloseHostShellInstance)
	s.Step(`^executing "minishift (.*)" in host shell$`,
		util.ExecuteMinishiftInHostShell)
	s.Step(`^executing "minishift (.*)" in host shell (succeeds|fails)$`,
		util.ExecuteMinishiftInHostShellSucceedsOrFails)
	s.Step(`^executing "(.*)" in host shell$`,
		util.ExecuteInHostShell)
	s.Step(`^executing "(.*)" in host shell (succeeds|fails)$`,
		util.ExecuteInHostShellSucceedsOrFails)

	// Shell output verification
	s.Step(`^with up to "(\d*)" retries with wait period of "(\d*)" (?:second|seconds) command "(.*)" output (?:should contain|contains) "(.*)"$`,
		util.ExecuteCommandInHostShellWithRetry)
	s.Step(`^(stdout|stderr) of host shell (?:should contain|contains) "(.*)"$`,
		util.HostShellCommandReturnShouldContain)
	s.Step(`^(stdout|stderr) of host shell (?:should not contain|does not contain) "(.*)"$`,
		util.HostShellCommandReturnShouldNotContain)
	s.Step(`^(stdout|stderr) of host shell (?:should contain|contains)$`,
		util.HostShellCommandReturnShouldContainContent)
	s.Step(`^(stdout|stderr) of host shell (?:should not contain|does not contain)$`,
		util.HostShellCommandReturnShouldNotContainContent)
	s.Step(`^(stdout|stderr) of host shell (?:should equal|equals) "(.*)"$`,
		util.HostShellCommandReturnShouldEqual)
	s.Step(`^(stdout|stderr) of host shell (?:should equal|equals)$`,
		util.HostShellCommandReturnShouldEqualContent)
	s.Step(`^evaluating stdout of the previous command in host shell succeeds$`,
		util.ExecuteInHostShellLineByLine)

	// Scenario variables
	// allows to set a scenario variable to the output values of minishift and oc commands
	// and then refer to it by $(NAME_OF_VARIABLE) directly in the text of feature file
	s.Step(`^setting scenario variable "(.*)" to the stdout from executing "oc (.*)"$`,
		MinishiftInstance.setVariableExecutingOcCommand)
	s.Step(`^setting scenario variable "(.*)" to the stdout from executing "minishift (.*)"$`,
		MinishiftInstance.setVariableExecutingMinishiftCommand)
	s.Step(`^scenario variable "(.*)" should not be empty$`,
		variableShouldNotBeEmpty)

	// Environment variables
	s.Step(`^setting up environment variable "(.*)" with value "(.*)" succeeds$`,
		setEnvironmentVariable)
	s.Step(`^unset environment variable "(.*)" succeeds$`,
		unSetEnvironmentVariable)

	// Image caching operations
	s.Step(`^image caching is (disabled|enabled)$`,
		MinishiftInstance.setImageCaching)
	s.Step(`^image export completes with (\d+) images within (\d+) minutes$`,
		MinishiftInstance.imageExportShouldComplete)
	s.Step(`^container image "(.*)" is cached$`,
		MinishiftInstance.imageShouldHaveCached)

	// Service rollout
	// to wait until service is deployed and ready before followin steps are started
	s.Step(`^services? "([^"]*)" rollout successfully$`,
		MinishiftInstance.rolloutServicesSuccessfully)
	s.Step(`^services? "([^"]*)" rollout successfully within "(\d+)" seconds$`,
		MinishiftInstance.rolloutServicesSuccessfullyBeforeTimeout)

	// Proxy testing
	// starts and sets a proxy server, checks for traffic on it, stops and unsets the server
	s.Step(`^user starts proxy server and sets MINISHIFT_HTTP_PROXY variable$`,
		testProxy.SetProxy)
	s.Step(`^user stops proxy server and unsets MINISHIFT_HTTP_PROXY variable$`,
		testProxy.UnsetProxy)
	s.Step(`^proxy log should contain "(.*)"$`,
		proxyLogShouldContain)
	s.Step(`^proxy log should contain$`,
		proxyLogShouldContainContent)

	// HTTP requests to OpenShift instance
	// to check OpenShift console and HTTP endpoint of deployed applications
	s.Step(`^"(body|status code)" of HTTP request to "([^"]*)" (contains|is equal to) "(.*)"$`,
		verifyRequestToURL)
	s.Step(`^with up to "(\d*)" retries with wait period of "(\d*)ms" the "(body|status code)" of HTTP request to "([^"]*)" (contains|is equal to) "(.*)"$`,
		verifyRequestToURLWithRetry)
	s.Step(`^"(body|status code)" of HTTP request to "([^"]*)" of OpenShift instance (contains|is equal to) "(.*)"$`,
		verifyRequestToOpenShift)
	s.Step(`^with up to "(\d*)" retries with wait period of "(\d*)ms" the "(body|status code)" of HTTP request to "([^"]*)" of OpenShift instance (contains|is equal to) "(.*)"$`,
		verifyRequestToOpenShiftWithRetry)
	s.Step(`^"(body|status code)" of HTTP request to "([^"]*)" of service "([^"]*)" in namespace "([^"]*)" (contains|is equal to) "(.*)"$`,
		verifyRequestToService)
	s.Step(`^with up to "(\d*)" retries with wait period of "(\d*)ms" the "(body|status code)" of HTTP request to "([^"]*)" of service "([^"]*)" in namespace "([^"]*)" (contains|is equal to) "(.*)"$`,
		verifyRequestToServiceWithRetry)

	// Config file content, JSON and YAML
	s.Step(`^(JSON|YAML) config file "(.*)" (contains|does not contain) key "(.*)" with value matching "(.*)"$`,
		configFileContainsKeyMatchingValue)
	s.Step(`^(JSON|YAML) config file "(.*)" (has|does not have) key "(.*)"$`,
		configFileContainsKey)
	s.Step(`^(stdout|stderr) is (JSON|YAML) which (contains|does not contain) key "(.*)" with value matching "(.*)"$`,
		stdoutContainsKeyMatchingValue)
	s.Step(`^(stdout|stderr) is (JSON|YAML) which (has|does not have) key "(.*)"$`,
		stdoutContainsKey)

	// Container status
	s.Step(`^with up to "(\d*)" retries with wait period of "(\d*)" (?:second|seconds) container name "(.*)" should be "(running|exited)"$`,
		MinishiftInstance.containerStatus)

	// Resource and config checks inside the running VM
	s.Step(`^Minishift VM should run with "(\d+)" vCPUs`,
		MinishiftInstance.ShouldHaveNoOfProcessors)
	s.Step(`^Minishift VM should run within "(\d+)" to "(\d+)" GB of disk size`,
		MinishiftInstance.ShouldHaveDiskSize)
	s.Step(`^printing Docker daemon configuration to stdout$`,
		catDockerConfigFile)

	s.Step(`^adding hostfolder of type "([^"]*)" of source directory "([^"]*)" to mount point "([^"]*)" of share name "([^"]*)" succeeds$`,
		MinishiftInstance.addHostFolder)
	s.Step(`^hostfolder share name "([^"]*)" (should|should not) be mounted$`, MinishiftInstance.hostFolderMountStatus)

	// File download
	// when external file is needed, for example, when downloading addon from minishift-addons
	s.Step(`^file from "(.*)" is downloaded into location "(.*)"$`,
		downloadFileIntoLocation)

	// Directory operations
	s.Step(`^deleting directory "([^"]*)" succeeds$`,
		deletingDirectorySucceeds)
	s.Step(`^directory "([^"]*)" shouldn\'t exist$`,
		directoryShouldntExist)
	s.Step(`^creating directory "([^"]*)" succeeds$`,
		creatingDirectoryInTestDirSucceeds)
	s.Step(`^creating file "([^"]*)" in directory "([^"]*)" succeeds$`,
		createFileInTestDirSucceeds)
	s.Step(`^writing text "([^"]*)" to file "([^"]*)" in directory path "([^"]*)" succeeds$`,
		writeToFileInTestDirSucceeds)
	s.Step(`^file "([^"]*)" should match text "([^"]*)" succeeds$`,
		readFileForTextMatchSucceeds)
	// Prototyping and debugging
	// please do not use in production
	s.Step(`^user (?:waits|waited) "(\d+)" seconds?$`,
		func(seconds int) error {
			time.Sleep(time.Duration(seconds) * time.Second)
			return nil
		})

	s.BeforeSuite(func() {
		testDir, testDefaultHome = setupTestDirectory()
		MinishiftInstance.runner.EnsureAllMinishiftHomesDeleted(testDir)
		SetMinishiftHomeEnv(testDefaultHome)
		ensureTestDirEmpty()

		testResultDir = filepath.Join(testDir, "..", "test-results")
		err := os.MkdirAll(testResultDir, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directory for test results:", err)
			os.Exit(1)
		}

		err = util.StartLog(testResultDir)
		if err != nil {
			fmt.Println("Error starting the log:", err)
			os.Exit(1)
		}

		fmt.Printf("Running Integration test in:%v.\nUsing MINISHIFT_HOME=%v\n", testDir, testDefaultHome)
		fmt.Println("Using binary:", minishiftBinary)
	})

	s.AfterSuite(func() {
		util.LogMessage("info", "----- Cleaning Up -----")
		MinishiftInstance.runner.EnsureAllMinishiftHomesDeleted(testDir)
		err := util.CloseLog()
		if err != nil {
			fmt.Println("Error closing the log:", err)
		}
	})

	s.BeforeFeature(func(this *gherkin.Feature) {
		util.LogMessage("info", "----- Preparing for feature -----")
		if runner.IsCDK() {
			runner.CDKSetup()
		} else {
			runner.RunCommandAndPrintError("addons list")
		}

		if runBeforeFeature != "" {
			err := runCommandsBeforeFeature(runBeforeFeature, &runner)
			fmt.Println("Running commands:", runBeforeFeature)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		if copyOcFrom != "" {
			err := copyOc(copyOcFrom, testDefaultHome)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		util.LogMessage("info", fmt.Sprintf("----- Feature: %s -----", this.Name))

		os.Chdir(testDir)
	})

	s.AfterFeature(func(this *gherkin.Feature) {
		util.LogMessage("info", "----- Cleaning after feature -----")
		cleanTestDefaultHomeConfiguration()
	})

	s.BeforeScenario(func(this interface{}) {
		switch this.(type) {
		case *gherkin.Scenario:
			scenario := *this.(*gherkin.Scenario)
			util.LogMessage("info", fmt.Sprintf("----- Scenario: %s -----", scenario.ScenarioDefinition.Name))
		case *gherkin.ScenarioOutline:
			scenario := *this.(*gherkin.ScenarioOutline)
			util.LogMessage("info", fmt.Sprintf("----- Scenario Outline: %s -----", scenario.ScenarioDefinition.Name))
		}
	})

	s.AfterScenario(func(interface{}, error) {
		testProxy.ResetLog(false)
	})

}

//  To get values of nested keys, use following dot formating in Scenarios: key.nestedKey
//  If an array is expected, then expect: "[value1 value2 value3]"
//  If empty string, non existing value are expected, then expect "<nil>"
func getConfigKeyValue(configData []byte, format string, keyPath string) (string, error) {
	var err error
	var keyValue string
	var values map[string]interface{}

	if format == "JSON" {
		err = json.Unmarshal(configData, &values)
		if err != nil {
			return "", fmt.Errorf("Error unmarshaling JSON: %s", err)
		}
	} else if format == "YAML" {
		err = yaml.Unmarshal(configData, &values)
		if err != nil {
			return "", fmt.Errorf("Error unmarshaling YAML: %s", err)
		}
	}

	keyPathArray := strings.Split(keyPath, ".")
	for _, element := range keyPathArray {
		switch value := values[element].(type) {
		case map[string]interface{}:
			values = value
		case map[interface{}]interface{}:
			retypedValue := make(map[string]interface{})
			for x := range value {
				retypedValue[x.(string)] = value[x]
			}
			values = retypedValue
		case []interface{}, nil, string, int, float64, bool:
			keyValue = fmt.Sprintf("%v", value)
		default:
			return "", errors.New("Unexpected type in config file, type not supported.")
		}
	}
	return keyValue, nil
}

func getFileContent(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(testDir + "/" + path)
	if err != nil {
		return nil, fmt.Errorf("Cannot read file: %v", err)
	}

	return data, err
}

func configFileContainsKeyMatchingValue(format string, configPath string, condition string, keyPath string, expectedValue string) error {
	config, err := getFileContent(configPath)
	if err != nil {
		return err
	}

	keyValue, err := getConfigKeyValue(config, format, keyPath)
	if err != nil {
		return err
	}

	matches, err := util.PerformRegexMatch(expectedValue, keyValue)
	if err != nil {
		return err
	} else if (condition == "contains") && !matches {
		return fmt.Errorf("For key '%s' config contains unexpected value '%s'", keyPath, keyValue)
	} else if (condition == "does not contain") && matches {
		return fmt.Errorf("For key '%s' config contains value '%s', which it should not contain", keyPath, keyValue)
	}

	return nil
}

func configFileContainsKey(format string, configPath string, condition string, keyPath string) error {
	config, err := getFileContent(configPath)
	if err != nil {
		return err
	}

	keyValue, err := getConfigKeyValue(config, format, keyPath)
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

func stdoutContainsKeyMatchingValue(commandField string, format string, condition string, keyPath string, expectedValue string) error {
	config := []byte(selectFieldFromLastOutput(commandField))

	keyValue, err := getConfigKeyValue(config, format, keyPath)
	if err != nil {
		return err
	}

	matches, err := util.PerformRegexMatch(expectedValue, keyValue)
	if err != nil {
		return err
	} else if (condition == "contains") && !matches {
		return fmt.Errorf("For key '%s' %s contains unexpected value '%s'", keyPath, commandField, keyValue)
	} else if (condition == "does not contain") && matches {
		return fmt.Errorf("For key '%s' %s contains value '%s', which it should not contain", keyPath, commandField, keyValue)
	}

	return nil
}

func stdoutContainsKey(commandField string, format string, condition string, keyPath string) error {
	config := []byte(selectFieldFromLastOutput(commandField))

	keyValue, err := getConfigKeyValue(config, format, keyPath)
	if err != nil {
		return err
	}

	if (condition == "has") && (keyValue == "<nil>") {
		return fmt.Errorf("%s does not contain any value for key %s", commandField, keyPath)
	} else if (condition == "does not have") && (keyValue != "<nil>") {
		return fmt.Errorf("%s contains key %s with assigned value: %s", commandField, keyPath, keyValue)
	}

	return nil
}

func GetLastCommandOutput() CommandOutput {
	return commandOutputs[len(commandOutputs)-1]
}

func selectFieldFromLastOutput(commandField string) string {
	lastCommandOutput := GetLastCommandOutput()
	outputField := ""
	switch commandField {
	case "stdout":
		outputField = lastCommandOutput.StdOut
	case "stderr":
		outputField = lastCommandOutput.StdErr
	case "exitcode":
		outputField = strconv.Itoa(lastCommandOutput.ExitCode)
	}
	return outputField
}

func shouldBeInValidFormat(commandField string, format string) error {
	result := selectFieldFromLastOutput(commandField)
	result = strings.TrimRight(result, "\n")
	var err error
	switch format {
	case "URL":
		_, err = validateURL(result)
	case "IP":
		_, err = validateIP(result)
	case "IP with port number":
		_, err = validateIPWithPort(result)
	case "YAML":
		_, err = validateYAML(result)
	default:
		return fmt.Errorf("Format %s not implemented.", format)
	}

	return err
}

func validateIP(inputString string) (bool, error) {
	if net.ParseIP(inputString) == nil {
		return false, fmt.Errorf("IP address '%s' is not a valid IP address", inputString)
	}

	return true, nil
}

func validateURL(inputString string) (bool, error) {
	_, err := url.ParseRequestURI(inputString)
	if err != nil {
		return false, fmt.Errorf("URL '%s' is not an URL in valid format. Parsing error: %v", inputString, err)
	}

	return true, nil
}

func validateIPWithPort(inputString string) (bool, error) {
	split := strings.Split(inputString, ":")
	if len(split) != 2 {
		return false, fmt.Errorf("String '%s' does not contain one ':' separator", inputString)
	}
	if _, err := strconv.Atoi(split[1]); err != nil {
		return false, fmt.Errorf("Port must be an integer, in '%s' the port '%s' is not an integer. Conversion error: %v", inputString, split[1], err)
	}
	if net.ParseIP(split[0]) == nil {
		return false, fmt.Errorf("In '%s' the IP part '%s' is not a valid IP address", inputString, split[0])
	}

	return true, nil
}

func validateYAML(inputString string) (bool, error) {
	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal([]byte(inputString), &m)
	if err != nil {
		return false, fmt.Errorf("Error unmarshaling YAML: %s. YAML='%s'", err, inputString)
	}

	return true, nil
}

func commandReturnEquals(commandField string, command string, condition string, expected string) error {
	MinishiftInstance.ExecutingMinishiftCommand(command)
	if condition == "is equal" {
		return util.CompareExpectedWithActualEquals(expected+"\n", selectFieldFromLastOutput(commandField))
	} else {
		return util.CompareExpectedWithActualNotEquals(expected+"\n", selectFieldFromLastOutput(commandField))
	}
}

func commandReturnContains(commandField string, command string, condition string, expected string) error {
	MinishiftInstance.ExecutingMinishiftCommand(command)
	if condition == "contains" {
		return util.CompareExpectedWithActualContains(expected, selectFieldFromLastOutput(commandField))
	} else {
		return util.CompareExpectedWithActualNotContains(expected, selectFieldFromLastOutput(commandField))
	}
}

func commandReturnShouldContain(commandField string, expected string) error {
	return util.CompareExpectedWithActualContains(expected, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldNotContain(commandField string, notexpected string) error {
	return util.CompareExpectedWithActualNotContains(notexpected, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldContainContent(commandField string, expected *gherkin.DocString) error {
	return util.CompareExpectedWithActualContains(expected.Content, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldNotContainContent(commandField string, notexpected *gherkin.DocString) error {
	return util.CompareExpectedWithActualNotContains(notexpected.Content, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldEqual(commandField string, expected string) error {
	return util.CompareExpectedWithActualEquals(expected, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldEqualContent(commandField string, expected *gherkin.DocString) error {
	return util.CompareExpectedWithActualEquals(expected.Content, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldBeEmpty(commandField string) error {
	return util.CompareExpectedWithActualEquals("", selectFieldFromLastOutput(commandField))
}

func commandReturnShouldNotBeEmpty(commandField string) error {
	return util.CompareExpectedWithActualNotEquals("", selectFieldFromLastOutput(commandField))
}

func variableShouldNotBeEmpty(variableName string) error {
	return util.CompareExpectedWithActualNotEquals("", util.GetVariableByName(variableName))
}

func commandReturnShouldMatchRegex(commandField string, expected string) error {
	return util.CompareExpectedWithActualMatchesRegex(expected, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldNotMatchRegex(commandField string, notexpected string) error {
	return util.CompareExpectedWithActualNotMatchesRegex(notexpected, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldMatchRegexContent(commandField string, expected *gherkin.DocString) error {
	return util.CompareExpectedWithActualMatchesRegex(expected.Content, selectFieldFromLastOutput(commandField))
}

func commandReturnShouldNotMatchRegexContent(commandField string, notexpected *gherkin.DocString) error {
	return util.CompareExpectedWithActualNotMatchesRegex(notexpected.Content, selectFieldFromLastOutput(commandField))
}

type commandRunner func(string) error

func ExecutingOcCommandSucceedsOrFails(command string, expectedResult string) error {
	return succeedsOrFails(MinishiftInstance.ExecutingOcCommand, command, expectedResult)
}

func ExecutingMinishiftCommandSucceedsOrFails(command string, expectedResult string) error {
	return succeedsOrFails(MinishiftInstance.ExecutingMinishiftCommand, command, expectedResult)
}

func succeedsOrFails(execute commandRunner, command string, expectedResult string) error {
	err := execute(command)
	if err != nil {
		return err
	}

	lastCommandOutput := GetLastCommandOutput()
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

func verifyRequestToURL(partOfResponse string, url string, assertion string, expected string) error {
	return verifyHTTPResponse(partOfResponse, url, assertion, expected)
}

func verifyRequestToURLWithRetry(retryCount int, retryWaitPeriod int, partOfResponse string, url string, assertion string, expected string) error {
	return verifyHTTPResponseWithRetry(partOfResponse, url, assertion, expected, retryCount, retryWaitPeriod)
}

func verifyRequestToService(partOfResponse string, urlSuffix string, serviceName string, nameSpace string, assertion string, expected string) error {
	url := MinishiftInstance.getRoute(serviceName, nameSpace) + urlSuffix
	return verifyHTTPResponse(partOfResponse, url, assertion, expected)
}

func verifyRequestToServiceWithRetry(retryCount int, retryWaitPeriod int, partOfResponse string, urlSuffix string, serviceName string, nameSpace string, assertion string, expected string) error {
	url := MinishiftInstance.getRoute(serviceName, nameSpace) + urlSuffix
	return verifyHTTPResponseWithRetry(partOfResponse, url, assertion, expected, retryCount, retryWaitPeriod)
}

func verifyRequestToOpenShift(partOfResponse string, urlSuffix string, assertion string, expected string) error {
	url := MinishiftInstance.getOpenShiftUrl() + urlSuffix
	return verifyHTTPResponse(partOfResponse, url, assertion, expected)
}

func verifyRequestToOpenShiftWithRetry(retryCount int, retryWaitPeriod int, partOfResponse string, urlSuffix string, assertion string, expected string) error {
	url := MinishiftInstance.getOpenShiftUrl() + urlSuffix
	return verifyHTTPResponseWithRetry(partOfResponse, url, assertion, expected, retryCount, retryWaitPeriod)
}

func verifyHTTPResponseWithRetry(partOfResponse string, url string, assertion string, expected string, retryCount int, retryWaitPeriod int) error {
	var err error

	for i := 0; i <= retryCount; i++ {
		err = verifyHTTPResponse(partOfResponse, url, assertion, expected)
		if err == nil {
			break
		}
		if i == retryCount {
			fmt.Printf("HTTP check (%v) has failed. Error: %v\n", i+1, err)
			break
		}

		fmt.Printf("HTTP check (%v) has failed, trying again in %v ms. Error: %v\n", i+1, retryWaitPeriod, err)
		time.Sleep(time.Millisecond * time.Duration(retryWaitPeriod))
	}

	return err
}

func verifyHTTPResponse(partOfResponse string, url string, assertion string, expected string) error {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}
	response, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("On url: '%v' the server returned an error: '%v'.", url, err)
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

func proxyLogShouldContain(expected string) error {
	return util.CompareExpectedWithActualContains(expected, testProxy.GetLog())
}

func proxyLogShouldContainContent(expected *gherkin.DocString) error {
	return util.CompareExpectedWithActualContains(expected.Content, testProxy.GetLog())
}

func catDockerConfigFile() error {
	var err error
	switch isoName {
	case "b2d":
		err = ExecutingMinishiftCommandSucceedsOrFails("ssh -- cat /var/lib/boot2docker/profile", "succeeds")
	case "centos", "rhel":
		err = ExecutingMinishiftCommandSucceedsOrFails("ssh -- cat /etc/systemd/system/docker.service.d/10-machine.conf", "succeeds")
	default:
		err = errors.New("ISO name not supported.")
	}

	return err
}

func downloadFileIntoLocation(downloadURL string, destinationFolder string) error {
	destinationFolder = filepath.Join(testDir, destinationFolder)
	err := os.MkdirAll(destinationFolder, os.ModePerm)
	if err != nil {
		return err
	}

	slice := strings.Split(downloadURL, "/")
	fileName := slice[len(slice)-1]
	filePath := filepath.Join(destinationFolder, fileName)
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func startHostShellInstance() error {
	return util.StartHostShellInstance(testWithShell, minishiftBinary)
}

func deletingDirectorySucceeds(dir string) error {
	return os.RemoveAll(dir)
}

func directoryShouldntExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	return fmt.Errorf("Directory %s exists", dir)
}

func setEnvironmentVariable(key string, value string) error {
	return os.Setenv(key, value)
}

func unSetEnvironmentVariable(key string) error {
	return os.Unsetenv(key)
}

func creatingDirectoryInTestDirSucceeds(dirName string) error {
	return os.MkdirAll(testDir+"/"+dirName, 0777)
}

func createFileInTestDirSucceeds(fileName string, filePath string) error {
	_, err := os.Stat(filePath + "/" + fileName)
	if os.IsNotExist(err) {
		file, err := os.Create(filePath + "/" + fileName)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

func writeToFileInTestDirSucceeds(text string, fileName string, filePath string) error {
	file, err := os.OpenFile(filePath+"/"+fileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(text)
	if err != nil {
		return err
	}
	err = file.Sync()
	if err != nil {
		return err
	}
	return nil
}

func readFileForTextMatchSucceeds(filePath string, textMatch string) error {
	text, err := getFileContent(filePath)
	if err != nil {
		return err
	}
	if strings.Trim(string(text), " \n") != textMatch {
		return fmt.Errorf("Expected: %s, Actual: %s", textMatch, text)
	}
	return nil
}
