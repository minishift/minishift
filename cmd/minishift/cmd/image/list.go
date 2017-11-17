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
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/spf13/cobra"
)

var (
	dockerDaemonImages bool

	imageCacheListCmd = &cobra.Command{
		Use:   "list ",
		Short: "Displays the locally cached images.",
		Long:  "Displays the locally cached images.",
		Run:   listImages,
	}
)

func listImages(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer api.Close()

	if dockerDaemonImages {
		listDockerDaemonImages(api)
	} else {
		listCachedImages()
	}
}

func listCachedImages() {
	cacheDir := state.InstanceDirs.ImageCache
	cachedImages := getCachedImages(cacheDir)
	if len(cachedImages) > 0 {
		printImageList(cachedImages)
	}
}

func listDockerDaemonImages(api *libmachine.Client) {
	images := getDockerDaemonImages(api)
	if len(images) == 0 {
		fmt.Println(fmt.Sprintf("There are no images available in the Docker daemon of Minishift instance '%s'", constants.ProfileName))
	} else {
		printImageList(images)
	}
}

func printImageList(images []string) {
	for _, i := range images {
		fmt.Println(i)
	}
}

func init() {
	imageCacheListCmd.Flags().BoolVar(&dockerDaemonImages, "vm", false, "Prints the available images in the Docker daemon.")
	ImageCmd.AddCommand(imageCacheListCmd)
}
