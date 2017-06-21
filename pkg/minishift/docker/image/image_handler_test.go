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

package image

import (
	"github.com/minishift/minishift/pkg/minikube/tests"
	"testing"
)

var testHandler *DockerImageHandler

type imageTestCase struct {
	fileName string
	isImage  bool
}

func Test_converting_image_name_to_file_name(t *testing.T) {
	setUp(t)

	actualFileName := testHandler.imageNameToFileName("openshift/origin:v1.5.1")
	expectedFileName := "openshift@origin@v1.5.1"
	if actualFileName != expectedFileName {
		t.Fatalf("Expected '%s', but got '%s'", expectedFileName, actualFileName)
	}
}

func Test_converting_file_name_to_image_name(t *testing.T) {
	setUp(t)

	actualImageName := testHandler.fileNameToImageName("/foo/bar/openshift@origin@v1.5.1")
	expectedImageName := "openshift/origin:v1.5.1"
	if actualImageName != expectedImageName {
		t.Fatalf("Expected '%s', but got '%s'", expectedImageName, actualImageName)
	}
}

func Test_is_image_name(t *testing.T) {
	setUp(t)

	testData := []imageTestCase{
		{fileName: "openshift@origin@v1.5.1", isImage: true},
		{fileName: "openshift@origin:v1.5.1", isImage: false},
		{fileName: "openshift/origin:v1.5.1", isImage: false},
	}

	for _, testCase := range testData {
		isImage := testHandler.isImageFile(testCase.fileName)
		if testCase.isImage != isImage {
			t.Fatalf("Expected isImageFile for '%s' to return %b, but got %b", testCase.fileName, testCase.isImage, isImage)
		}

	}
}

func setUp(t *testing.T) {
	var err error
	testHandler, err = NewDockerImageHandler(&tests.MockDriver{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
