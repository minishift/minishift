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
	"os"
	"path/filepath"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/docker/image"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
)

var (
	logToFile bool
	exportAll bool

	imageExportCmd = &cobra.Command{
		Use:   "export [image ...]",
		Short: "Exports the specified container images.",
		Long:  "Exports the specified container images.",
		Run:   exportImage,
	}
)

const (
	noDockerDaemonImages = "There are currently no images in the Docker daemon which can be exported."
)

func exportImage(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer api.Close()

	util.ExitIfUndefined(api, constants.MachineName)

	host, err := api.Load(constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating the VM client: %v", err))
	}

	util.ExitIfNotRunning(host.Driver, constants.MachineName)

	envMap, err := cluster.GetHostDockerEnv(api)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error determining Docker daemon settings: %v", err))
	}

	var out io.Writer
	if logToFile {
		logFile := createLogFile()
		defer logFile.Close()
		out = logFile
	} else {
		out = os.Stdout
	}

	images := imagesToExport(api, args)

	handler, err := image.NewOciImageHandler(host.Driver, envMap)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Cannot create the image handler: %v", err))
	}

	normalizedImageNames, err := normalizeImageNames(images)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("%v contains an invalid image names:\n%v", images, err.Error()))
	}

	imageCacheConfig := &image.ImageCacheConfig{
		HostCacheDir:      state.InstanceDirs.ImageCache,
		CachedImages:      normalizedImageNames,
		Out:               out,
		ImageMissStrategy: image.Pull,
	}

	err = handler.ExportImages(imageCacheConfig)
	if err != nil {
		msg := fmt.Sprintf("Container image export failed:\n%v", err)
		if logToFile {
			fmt.Fprint(out, msg)
		}
		atexit.ExitWithMessage(1, msg)
	}
}

func imagesToExport(api *libmachine.Client, args []string) []string {
	var images []string
	if exportAll {
		images = getDockerDaemonImages(api)
	} else if len(args) == 0 {
		images = viper.GetStringSlice(config.CacheImages.Name)
	} else {
		images = args
	}

	if len(images) == 0 {
		msg := noCachedImagesSpecified
		if importAll {
			msg = noDockerDaemonImages
		}
		atexit.ExitWithMessage(0, msg)
	}

	return images

}

func createLogFile() *os.File {
	now := time.Now()
	timeStamp := now.Format("2006-01-02-1504-05") // reference time Mon Jan 2 15:04:05 -0700 MST 2006
	logFilePath := filepath.Join(state.InstanceDirs.Logs, fmt.Sprintf("image-export-%s.log", timeStamp))
	logFile, err := os.Create(logFilePath)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Cannot create the log file of the image export: %v", err))
	}

	return logFile
}

func init() {
	imageExportCmd.Flags().BoolVar(&exportAll, "all", false, "Exports all images currently available in the Docker daemon.")
	imageExportCmd.Flags().BoolVar(&logToFile, "log-to-file", false, "Logs export progress to file instead of standard out.")
	ImageCmd.AddCommand(imageExportCmd)
}
