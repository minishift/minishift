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
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/docker/image"
	"github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var (
	imageCacheListCmd = &cobra.Command{
		Use:   "list ",
		Short: "Displays the locally cached images (experimental).",
		Long:  "Displays the locally cached images (experimental).",
		Run:   listImages,
	}
)

func listImages(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	handler, err := image.NewLocalOnlyOciImageHandler()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Cannot create the image handler: %v", err))
	}

	cacheDir := constants.MakeMiniPath(CacheDir...)
	imageCacheConfig := &image.ImageCacheConfig{
		HostCacheDir:      cacheDir,
		CachedImages:      args,
		Out:               os.Stdout,
		ImageMissStrategy: image.Pull,
	}

	images := handler.GetCachedImages(imageCacheConfig)
	if len(images) == 0 {
		fmt.Println(fmt.Sprintf("There are no images cached for profile '%s' in in '%s'", profile.GetActiveProfile(), cacheDir))
	} else {
		for k := range images {
			fmt.Println(k)
		}
	}
}

func init() {
	ImageCmd.AddCommand(imageCacheListCmd)
}
