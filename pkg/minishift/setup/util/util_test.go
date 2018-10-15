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
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test_folder_contains_filenames(t *testing.T) {
	testFiles := []string{"foo.exe", "bar.exe"}
	testDir, _ := ioutil.TempDir("", "minishift-folder-containstest-")
	tmpFile, err := os.Create(filepath.Join(testDir, testFiles[0]))
	assert.NoError(t, err)
	tmpFile, err = os.Create(filepath.Join(testDir, testFiles[1]))
	assert.NoError(t, err)
	defer os.RemoveAll(testDir)

	fmt.Println("tmpFile: ", tmpFile.Name())
	assert.Truef(t, FolderContains(testDir, testFiles),
		fmt.Sprintf("Folder '%s' should contain every files of '%v'", testDir, testFiles))
}

func Test_folder_contains_invalid_filenames(t *testing.T) {
	testFiles := []string{"foo.exe", "bar.exe"}
	testDir, _ := ioutil.TempDir("", "minishift-folder-containstest-")
	_, err := os.Create(filepath.Join(testDir, testFiles[0]))
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(testDir, testFiles[1]))
	assert.NoError(t, err)
	defer os.RemoveAll(testDir)

	assert.Falsef(t, FolderContains(testDir, []string{testFiles[0], "oooo.exe"}),
		fmt.Sprintf("Folder '%s' should contain every files of '%v'", testDir, testFiles))
}
