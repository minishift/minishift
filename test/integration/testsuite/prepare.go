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
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/test/integration/util"
)

const (
	delimiterConst = ";"
)

func ParseFlags() {
	flag.StringVar(&minishiftArgs, "minishift-args", "", "Arguments to pass to minishift")
	flag.StringVar(&minishiftBinary, "binary", "", "Path to minishift binary")
	flag.StringVar(&runBeforeFeature, "run-before-feature", "", "Set of minishift commands to be executed before every feature. Individual commands must be delimited by a semicolon.")
	flag.StringVar(&testWithShell, "test-with-specified-shell", "", "Name of shell to be used for steps which executes commands directly in persistent shell instance.")
	flag.StringVar(&copyOcFrom, "copy-oc-from", "", "Path to binary, the binary will $MINISHIFT_HOME/cache/oc.")
	flag.StringVar(&testDir, "test-dir", "", "Path to the directory in which to execute the tests")

	flag.StringVar(&GodogFormat, "format", "pretty", "Sets which format godog will use")
	flag.StringVar(&GodogTags, "tags", "", "Tags for godog test")
	flag.BoolVar(&GodogShowStepDefinitions, "definitions", false, "")
	flag.BoolVar(&GodogStopOnFailure, "stop-on-failure ", false, "Stop when failure is found")
	flag.BoolVar(&GodogNoColors, "no-colors", false, "Disable colors in godog output")
	flag.StringVar(&GodogPaths, "paths", "./features", "")

	flag.Parse()
}

func HandleISOVersion() {
	if GodogTags != "" {
		GodogTags += "&&"
	}
	runner := util.MinishiftRunner{CommandPath: minishiftBinary}
	if runner.IsCDK() {
		GodogTags += "~minishift-only"
		isoName = "rhel"
		fmt.Println("Test run using CDK binary with RHEL iso.")
	} else {
		GodogTags += "~cdk-only"
		isoUrl := os.Getenv("MINISHIFT_ISO_URL")
		switch isoUrl {
		case "b2d":
			fmt.Println("Test run using Boot2Docker iso image.")
			isoName = "b2d"
		case "minikube":
			fmt.Println("Test run using Minikube iso image.")
			isoName = "minikube"
		case "", "centos":
			fmt.Println("Test run using CentOS iso image.")
			isoName = "centos"
		default:
			fmt.Print("Using full path for iso image.")
			isoName = determineIsoFromFile(isoUrl)
		}
	}
}

func determineIsoFromFile(isoUrl string) string {
	var isoName string
	if matched, _ := regexp.MatchString(".*b2d\\.iso", isoUrl); matched {
		fmt.Println("Boot2docker variant was assumed from the filename of ISO.")
		isoName = "b2d"
	} else if matched, _ := regexp.MatchString(".*minikube\\.iso", isoUrl); matched {
		fmt.Println("Minikube variant was assumed from the filename of ISO.")
		isoName = "minikube"
	} else if matched, _ := regexp.MatchString(".*centos7\\.iso", isoUrl); matched {
		fmt.Println("CentOS variant was assumed from the filename of ISO.")
		isoName = "centos"
	} else {
		fmt.Println("Can't assume ISO variant from its filename. Will use Boot2Docker. To avoid this situation please name your ISO to end with 'b2d.iso', 'centos7.iso' or 'minikube.iso'.")
		isoName = "b2d"
	}
	return isoName
}

func setupTestDirectory() (string, string) {
	var err error
	if testDir == "" {
		testDir, err = ioutil.TempDir("", "minishift-integration-test-")
		if err != nil {
			fmt.Println("Error creating temporary directory:", err)
			os.Exit(1)
		}
	}
	testDefaultHome := fmt.Sprintf("%s/.minishift", testDir)

	return testDir, testDefaultHome
}

func SetMinishiftHomeEnv(path string) {
	err := os.Setenv(constants.MiniShiftHomeEnv, path)
	if err != nil {
		fmt.Printf("Error setting up environmental variable %v: %v\n", constants.MiniShiftHomeEnv, err)
	}
}

func ensureTestDirEmpty() {
	files, err := ioutil.ReadDir(testDir)
	if err != nil {
		fmt.Println(fmt.Sprintf("Unable to setup integration test directory: %v", err))
		os.Exit(1)
	}

	for _, file := range files {
		os.RemoveAll(filepath.Join(testDir, file.Name()))
	}
}

func cleanTestDefaultHomeConfiguration() {
	var foldersToClean []string
	foldersToClean = append(foldersToClean, filepath.Join(testDefaultHome, "addons"))
	foldersToClean = append(foldersToClean, filepath.Join(testDefaultHome, "config"))
	foldersToClean = append(foldersToClean, filepath.Join(testDefaultHome, "cache/images"))
	foldersToClean = append(foldersToClean, filepath.Join(testDefaultHome, "logs"))

	for index := range foldersToClean {
		err := os.RemoveAll(foldersToClean[index])
		if err != nil {
			fmt.Println(fmt.Sprintf("Unable to remove folder %v: %v", foldersToClean[index], err))
			os.Exit(1)
		}
	}
}

// runBeforeFeature executes minishift commands delimited by a semicolon before the feature starts.
func runCommandsBeforeFeature(commands string, runner *util.MinishiftRunner) error {
	var errorMessage string
	splittedCommands := strings.Split(commands, delimiterConst)
	fmt.Println("Going to run commands:", commands)

	for i := range splittedCommands {
		_, stdErr, exitCode := runner.RunCommand(splittedCommands[i])
		if exitCode != 0 {
			errorMessage += fmt.Sprintf("Error executing command 'minishift %v'.\nExit code: '%v',\nStderr: '%v'\n", splittedCommands[i], exitCode, stdErr)
		}
	}

	if errorMessage != "" {
		return errors.New(errorMessage)
	}

	return nil
}

// copyOc copies the oc binary contained in directory structure of /<version>/<platform>/oc(.exe)
// to cache/oc inside of the testDefaultHome.
func copyOc(ocPath string, testDefaultHome string) error {
	ocPath = filepath.Clean(ocPath)
	dir, ocFileName := filepath.Split(ocPath)
	dir, platformDir := filepath.Split(filepath.Clean(dir))
	_, versionDir := filepath.Split(filepath.Clean(dir))

	targetDirectory := filepath.Join(testDefaultHome, "cache", "oc", versionDir, platformDir)
	err := os.MkdirAll(targetDirectory, 0777)
	if err != nil {
		return fmt.Errorf("Error creating target directory:%v\n", err)
	}

	ocFile, err := os.Open(ocPath)
	if err != nil {
		return fmt.Errorf("Error accessing oc binary:%v\n", err)
	}
	defer ocFile.Close()

	targetOcPath := filepath.Join(targetDirectory, ocFileName)
	targetOc, err := os.Create(targetOcPath)
	if err != nil {
		return fmt.Errorf("oc binary target location cannot be opened:%v\n", err)
	}
	defer targetOc.Close()

	_, err = io.Copy(targetOc, ocFile)
	if err != nil {
		return fmt.Errorf("Error copying the binary:%v\n", err)
	}

	err = os.Chmod(targetOcPath, 0755)
	if err != nil {
		return fmt.Errorf("Error setting up the mode of oc binary:%v", err)
	}

	message := fmt.Sprintf("The oc binary has been copied from:'%v' into:'%v'.", ocPath, targetOcPath)
	util.LogMessage("info", message)
	fmt.Println(message)

	return nil
}
