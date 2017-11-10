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

// ImageHandler is responsible for the import and export of images into the Docker daemon of the VM
type ImageHandler interface {
	// ImportImages imports cached images from the host into the Docker daemon of the VM.
	ImportImages(config *ImageCacheConfig) error

	// ExportImages exports the images specified as part of the ImageCacheConfig from the VM to the host.
	ExportImages(config *ImageCacheConfig) error

	// IsImageCached returns true if the specified image is cached, false otherwise.
	IsImageCached(config *ImageCacheConfig, image string) bool

	// AreImagesCached returns true if all images specified in the config are cached, false otherwise.
	AreImagesCached(config *ImageCacheConfig) bool

	// GetCachedImages returns a map of cached image names. A map is used to make lookup for a specific image easier.
	GetCachedImages(config *ImageCacheConfig) map[string]bool
}
