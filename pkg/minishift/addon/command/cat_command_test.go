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

package command

import (
	"io/ioutil"
	"testing"

	pkgTesting "github.com/minishift/minishift/pkg/testing"
	"github.com/stretchr/testify/assert"
	"os"
)

func Test_cat_command_when_file_exist(t *testing.T) {

	// Creating a tmp file with test content.
	content := []byte("This file to used to test the cat command output.")
	tmpfile, err := ioutil.TempFile("", "test.txt")
	assert.NoError(t, err)

	defer os.Remove(tmpfile.Name()) // clean up

	_, err = tmpfile.Write(content)
	assert.NoError(t, err)

	err = tmpfile.Close()
	assert.NoError(t, err)

	testCases := []struct {
		catCommand     string
		expectedOutput string
		ignoreError    bool
	}{
		{
			catCommand:     tmpfile.Name(),
			expectedOutput: "This file to used to test the cat command output.",
			ignoreError:    false,
		},
	}

	for _, test := range testCases {
		tee, err := pkgTesting.NewTee(true)
		defer tee.Close()
		assert.NoError(t, err, "Error getting instance of Tee")

		cat := NewCatCommand(test.catCommand, test.ignoreError, "")
		context := &FakeInterpolationContext{}
		err = cat.Execute(&ExecutionContext{interpolationContext: context})
		assert.NoError(t, err)
		tee.Close()

		assert.Equal(t, tee.StdoutBuffer.String(), test.expectedOutput)
	}
}

func Test_cat_command_when_file_not_exist(t *testing.T) {
	testCases := []struct {
		catCommand     string
		expectedOutput string
		ignoreError    bool
		expectedError  string
	}{
		{
			catCommand:     "dummy",
			expectedOutput: "",
			ignoreError:    false,
			expectedError:  "File dummy doesn't exist",
		},
	}

	for _, test := range testCases {
		tee, err := pkgTesting.NewTee(true)
		defer tee.Close()
		assert.NoError(t, err, "Error getting instance of Tee")

		cat := NewCatCommand(test.catCommand, test.ignoreError, "")
		context := &FakeInterpolationContext{}
		err = cat.Execute(&ExecutionContext{interpolationContext: context})
		assert.EqualError(t, err, test.expectedError)
		tee.Close()
	}
}
