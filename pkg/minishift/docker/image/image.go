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
	"github.com/minishift/minishift/pkg/util/os"
	"github.com/minishift/minishift/pkg/util/os/process"
	"io"
	"os/exec"
)

type ImageMissStrategy int

const (
	Skip ImageMissStrategy = iota
	Pull
)

type ImageCacheConfig struct {
	HostCacheDir      string
	CachedImages      []string
	Out               io.Writer
	ImageMissStrategy ImageMissStrategy
}

// GetOpenShiftImageNames returns the full images names for the images requires for a fully functioning OpenShift instance
func GetOpenShiftImageNames(version string) []string {
	return []string{
		fmt.Sprintf("openshift/origin-control-plane:%s", version),
		fmt.Sprintf("openshift/origin-docker-registry:%s", version),
		fmt.Sprintf("openshift/origin-haproxy-router:%s", version),
	}
}

func CreateExportCommand(version string, profile string, images []string) (*exec.Cmd, error) {
	cmd, err := os.CurrentExecutable()
	if err != nil {
		return nil, err
	}

	exportArgs := []string{
		"--profile",
		profile,
		"image",
		"export",
		"--log-to-file",
	}
	exportArgs = append(exportArgs, images...)
	exportCmd := exec.Command(cmd, exportArgs...)
	// don't inherit any file handles
	exportCmd.Stderr = nil
	exportCmd.Stdin = nil
	exportCmd.Stdout = nil
	exportCmd.SysProcAttr = process.SysProcForBackgroundProcess()
	exportCmd.Env = process.EnvForBackgroundProcess()

	return exportCmd, nil
}
