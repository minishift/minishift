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
	"github.com/containers/image/docker/reference"
	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/docker/image"
	pkgUtil "github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/pkg/errors"
	"os"
	"sort"
)

func getCachedImages(cacheDir string) []string {
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	handler, err := image.NewLocalOnlyOciImageHandler()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Cannot create the image handler: %v", err))
	}

	imageCacheConfig := &image.ImageCacheConfig{
		HostCacheDir:      cacheDir,
		Out:               os.Stdout,
		ImageMissStrategy: image.Skip,
	}

	images := handler.GetCachedImages(imageCacheConfig)
	return sortImageNames(images)
}

func getDockerDaemonImages(api *libmachine.Client) []string {
	util.ExitIfUndefined(api, constants.MachineName)

	host, err := api.Load(constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
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

	images, err := handler.GetDockerImages()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error retrieving image list from Docker daemon: %v", err))
	}
	return sortImageNames(images)
}

func normalizeImageNames(images []string) ([]string, error) {
	mutliError := pkgUtil.MultiError{}
	normalizedImageNames := []string{}
	for _, image := range images {
		normalizedName, err := normalizeImageName(image)
		if err != nil {
			mutliError.Collect(errors.New(fmt.Sprintf("Error parsing image name '%s': %v", image, err)))
		}
		normalizedImageNames = append(normalizedImageNames, normalizedName)
	}
	return normalizedImageNames, mutliError.ToError()
}

func normalizeImageName(name string) (string, error) {
	ref, err := reference.Parse(name)
	if err != nil {
		return "", err
	}

	_, ok := ref.(reference.Tagged)
	if ok {
		return ref.String(), nil
	}

	ref, err = reference.WithTag(ref.(reference.Named), "latest")
	if err != nil {
		return "", err
	}

	return ref.String(), nil
}

func sortImageNames(images map[string]bool) []string {
	var sortedImageList []string
	for i := range images {
		sortedImageList = append(sortedImageList, i)
	}
	sort.Strings(sortedImageList)
	return sortedImageList
}

func toStringSlice(interfaceSlice []interface{}) []string {
	var slice []string
	for _, s := range interfaceSlice {
		slice = append(slice, s.(string))
	}
	return slice
}

func getConfiguredCachedImages(minishiftConfig config.MinishiftConfig) []string {
	var cacheImages []string
	if minishiftConfig[config.CacheImages.Name] == nil {
		cacheImages = []string{}
	} else {
		cacheImages = toStringSlice(minishiftConfig[config.CacheImages.Name].([]interface{}))
	}
	return cacheImages
}

func getMinishiftConfig() config.MinishiftConfig {
	minishiftConfig, err := config.ReadConfig()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Cannot read the Minishift configuration: %s", err.Error()))
	}
	return minishiftConfig
}
