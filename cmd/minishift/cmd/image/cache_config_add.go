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
	"github.com/spf13/cobra"

	"fmt"
	"github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/minishift/minishift/pkg/util/strings"
)

var (
	addConfiguredImageCmd = &cobra.Command{
		Use:   "add [image ...]",
		Short: "Adds the specified images to the list of configured images for import and export.",
		Long:  "Adds the specified images to the list of configured images for import and export.",
		Run:   addConfiguredImage,
	}
)

func addConfiguredImage(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		atexit.ExitWithMessage(1, noImageSpecified)
	}

	normalizedImageNames, err := normalizeImageNames(args)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Invalid image name: %v", err))
	}

	minishiftConfig := getMinishiftConfig()
	cacheImages := getConfiguredCachedImages(minishiftConfig)

	for _, image := range normalizedImageNames {
		if !strings.Contains(cacheImages, image) {
			cacheImages = append(cacheImages, image)
		}
	}

	minishiftConfig[config.CacheImages.Name] = cacheImages
	config.WriteConfig(minishiftConfig)
}

func init() {
	ImageCacheConfigCmd.AddCommand(addConfiguredImageCmd)
}
