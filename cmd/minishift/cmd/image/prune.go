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

package image

import (
	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minishift/docker/image"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var (
	pruneAll      bool
	imagePruneCmd = &cobra.Command{
		Use:     "delete [image ...]",
		Aliases: []string{"prune"},
		Short:   "Deletes the specified container images.",
		Long:    "Deletes the specified container images.",
		Run:     pruneImage,
	}
)

const (
	noImageProvided = "You need to either specify a list of images or a single image on the command line"
)

func pruneImage(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer api.Close()

	if pruneAll {
		fmt.Printf(fmt.Sprintf("Deleting all cached images from local cache ... "))
		if err := deleteCachedImages(state.InstanceDirs.ImageCache); err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Cannot delete the images from cache: %v", err))
		}
		fmt.Println("OK")
		return
	}

	images := imagesToPrune(api, args)

	handler, err := image.NewLocalOnlyOciImageHandler()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Cannot create the image handler: %v", err))
	}

	normalizedImageNames, err := normalizeImageNames(images)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("%v contains an invalid image names:\n%v", images, err.Error()))
	}

	imageCacheConfig := &image.ImageCacheConfig{
		HostCacheDir: state.InstanceDirs.ImageCache,
		CachedImages: normalizedImageNames,
	}

	_, err = handler.PruneImages(imageCacheConfig)
	if err != nil {
		msg := fmt.Sprintf("Deletion of the container image failed:\n%v", err)
		atexit.ExitWithMessage(1, msg)
	}
}

func imagesToPrune(api *libmachine.Client, args []string) []string {
	images := args
	if len(images) == 0 {
		atexit.ExitWithMessage(1, noImageProvided)
	}

	return images

}

func init() {
	imagePruneCmd.Flags().BoolVarP(&pruneAll, "all", "a", false, "Deletes all images available in the local image cache.")
	ImageCmd.AddCommand(imagePruneCmd)
}
