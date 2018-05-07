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
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"sort"
	"strings"
)

var (
	listConfiguredImagesCmd = &cobra.Command{
		Use:   "view",
		Short: "Displays the configured list of images for import and export.",
		Long:  "Displays the configured list of images for import and export.",
		Run:   viewConfiguredImages,
	}
)

func viewConfiguredImages(cmd *cobra.Command, args []string) {
	cacheImages := minishiftConfig.InstanceConfig.CacheImages

	if len(cacheImages) == 0 {
		return
	}

	sort.Strings(cacheImages)
	fmt.Println(strings.Join(cacheImages, "\n"))
}

func init() {
	ImageCacheConfigCmd.AddCommand(listConfiguredImagesCmd)
}
