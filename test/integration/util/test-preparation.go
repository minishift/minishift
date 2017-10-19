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

package util

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	delimiterConst = ";"
)

// RunBeforeFeature executes minishift commands delimited by a semicolon before the feature starts.
func RunBeforeFeature(commands string, runner *MinishiftRunner) error {
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

//CopyOc copies the oc binary contained in directory structure of /<version>/<platform>/oc(.exe)
//to cache/oc inside of the testDir.
func CopyOc(ocPath string, testDir string) error {
	ocPath = filepath.Clean(ocPath)
	dir, ocFileName := filepath.Split(ocPath)
	dir, platformDir := filepath.Split(filepath.Clean(dir))
	_, versionDir := filepath.Split(filepath.Clean(dir))

	targetDirectory := filepath.Join(testDir, "cache", "oc", versionDir, platformDir)
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
	LogMessage("info", message)
	fmt.Println(message)

	return nil
}
