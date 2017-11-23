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

package filehelper

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_non_existent_directory_returns_false(t *testing.T) {
	path := filepath.Join("this", "path", "really", "should", "not", "exists", "unless", "you", "have", "a", "crazy", "setup")
	assert.False(t, Exists(path), "The path '%s' should not exist", path)
}

func Test_existent_directory_returns_true(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-filetest-")
	defer os.RemoveAll(testDir)

	assert.NoError(t, err)

	assert.True(t, Exists(testDir))
}

func Test_testDir_is_directory(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-filetest-")
	defer os.RemoveAll(testDir)

	assert.NoError(t, err)

	assert.True(t, IsDirectory(testDir), "The path '%s' should be a directory", testDir)
}

func Test_non_existing_file_is_not_a_directory(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-filetest-")
	defer os.RemoveAll(testDir)

	assert.NoError(t, err)

	path := IsDirectory(filepath.Join(testDir, "foo"))
	assert.False(t, path, "The path '%s' should not be a directory", testDir)
}

func Test_file_is_not_a_directory(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-filetest-")
	defer os.RemoveAll(testDir)

	assert.NoError(t, err)

	content := []byte("Hello world")
	tmpfile, err := ioutil.TempFile(testDir, "example")
	defer tmpfile.Close()
	assert.NoError(t, err)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err)

	assert.False(t, IsDirectory(tmpfile.Name()), "The path '%s' should not be a directory", testDir)
}

func Test_non_existing_directory(t *testing.T) {
	testDir := "/foo/bar"
	empty := IsEmptyDir(testDir)
	assert.False(t, empty, "Expected that the directory %s doesn't exists", testDir)

}

func Test_existing_empty_directory(t *testing.T) {
	testDir, err := ioutil.TempDir("", "minishift-test-filetest-")
	defer os.RemoveAll(testDir)
	assert.NoError(t, err)

	empty := IsEmptyDir(testDir)
	assert.True(t, empty, "Expected  %s to be empty", testDir)
}

func Test_existing_nonempty_directory(t *testing.T) {
	testDir, _ := ioutil.TempDir("", "minishift-test-filetest-")
	_, err := ioutil.TempFile(testDir, "foo")
	defer os.RemoveAll(testDir)
	assert.NoError(t, err)

	empty := IsEmptyDir(testDir)
	assert.False(t, empty, "Expected %s to be nonempty.", testDir)
}
