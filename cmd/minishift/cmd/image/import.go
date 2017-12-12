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
)

var (
	importAll bool

	imageImportCmd = &cobra.Command{
		Use:   "import [image ...]",
		Short: "Imports the specified images into the Docker daemon.",
		Long:  "Imports the specified images into the Docker daemon.",
		Run:   importImage,
	}
)

const (
	noCachedImages = "There are currently no images in the local cache."
)

func importImage(cmd *cobra.Command, args []string) {
	cacheDir := state.InstanceDirs.ImageCache
	var images []string
	if importAll {
		images = getCachedImages(cacheDir)
	} else if len(args) == 0 {
		images = viper.GetStringSlice(config.CacheImages.Name)
	} else {
		images = args
	}

	if len(images) == 0 {
		msg := noCachedImagesSpecified
		if importAll {
			msg = noCachedImages
		}
		atexit.ExitWithMessage(0, msg)
	}

	normalizedImageNames, err := normalizeImageNames(images)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("%v contains an invalid image names:\n%v", images, err.Error()))
	}

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

	handler, err := image.NewOciImageHandler(host.Driver, envMap)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Cannot create the image handler: %v", err))
	}

	imageCacheConfig := &image.ImageCacheConfig{
		HostCacheDir:      state.InstanceDirs.ImageCache,
		CachedImages:      normalizedImageNames,
		Out:               os.Stdout,
		ImageMissStrategy: image.Skip,
	}

	importedImages, err := handler.ImportImages(imageCacheConfig)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Container image import failed:\n%v", err))
	}

	if len(importedImages) < len(normalizedImageNames) {
		atexit.ExitWithMessage(1, "At least one image could not be imported.")
	}
}

func init() {
	imageImportCmd.Flags().BoolVar(&importAll, "all", false, "Imports all images available in the local image cache.")
	ImageCmd.AddCommand(imageImportCmd)
}
