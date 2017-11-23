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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testData struct {
	image           string
	normalizedImage string
}

var tests = []testData{
	{image: "alpine", normalizedImage: "alpine:latest"},
	{image: "alpine:1.24", normalizedImage: "alpine:1.24"},
}

func Test_normalize_image_names(t *testing.T) {
	for _, test := range tests {
		normalizedImage, err := normalizeImageName(test.image)

		assert.NoError(t, err)
		assert.Equal(t, normalizedImage, test.normalizedImage, fmt.Sprintf("Normalizing '%s' should have returned '%s', got '%s'", test.image, normalizedImage, test.normalizedImage))
	}
}
