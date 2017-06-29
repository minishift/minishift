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
	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/docker/image"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

var imageExportCmd = &cobra.Command{
	Use:   "export [image ...]",
	Short: "Exports the specified container images (experimental).",
	Long:  "Exports the specified container images (experimental).",
	Run:   exportImage,
}

func exportImage(cmd *cobra.Command, args []string) {
	logFile := createLogFile()
	defer logFile.Close()

	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	if len(args) < 1 {
		atexit.ExitWithMessage(0, "You must specify at least one container image.")
	}

	util.ExitIfUndefined(api, constants.MachineName)

	host, err := api.Load(constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating the VM client: %v", err))
	}

	util.ExitIfNotRunning(host.Driver, constants.MachineName)

	handler, err := image.NewDockerImageHandler(host.Driver)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Cannot create the image handler: %v", err))
	}

	imageCacheConfig := &image.ImageCacheConfig{
		HostCacheDir:      constants.MakeMiniPath("cache", "images"),
		CachedImages:      args,
		Out:               logFile,
		ImageMissStrategy: image.PULL,
	}
	err = handler.ExportImages(imageCacheConfig)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Failed to export the container images: %v", err))
	}
}

func createLogFile() *os.File {
	now := time.Now()
	timeStamp := now.Format("2017-01-02-1504-00")
	logFilePath := filepath.Join(constants.MakeMiniPath("logs"), fmt.Sprintf("image-export-%s.log", timeStamp))
	logFile, err := os.Create(logFilePath)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Cannot create the log file of the image export: %v", err))
	}

	return logFile
}

func init() {
	ImageCmd.AddCommand(imageExportCmd)
}
