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
	"testing"

	"github.com/minishift/minishift/cmd/testing/cli"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

func Test_no_images_to_import(t *testing.T) {
	tee := cli.CreateTee(t, true)
	defer cli.TearDown("", tee)
	expectedOut := noCachedImagesSpecified

	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 0, expectedOut))
	importImage(nil, nil)
}

func Test_no_images_cached(t *testing.T) {
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)
	importAll = true

	expectedOut := noCachedImages

	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 0, expectedOut))
	importImage(nil, nil)
}
