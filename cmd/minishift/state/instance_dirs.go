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

package state

import (
	"path/filepath"

	"github.com/minishift/minishift/pkg/minikube/constants"
)

type MinishiftDirs struct {
	Home         string
	Config       string
	GlobalConfig string
	Machines     string
	Certs        string
	Cache        string
	IsoCache     string
	OcCache      string
	ImageCache   string
	Addons       string
	Logs         string
	Tmp          string
}

var InstanceDirs *MinishiftDirs

// Constructor for MinishiftDirs
func NewMinishiftDirs(baseDir string) *MinishiftDirs {
	minishiftHomeDir := constants.GetMinishiftHomeDir()
	// We use a global cache, sharing the cache directory of the default 'minishift' instance
	cacheDir := filepath.Join(minishiftHomeDir, "cache")

	return &MinishiftDirs{
		Home:         baseDir,
		Certs:        filepath.Join(baseDir, "certs"),
		Machines:     filepath.Join(baseDir, "machines"),
		Addons:       filepath.Join(baseDir, "addons"),
		Logs:         filepath.Join(baseDir, "logs"),
		Tmp:          filepath.Join(baseDir, "tmp"),
		Config:       filepath.Join(baseDir, "config"),
		GlobalConfig: filepath.Join(minishiftHomeDir, "config"),
		Cache:        cacheDir,
		IsoCache:     filepath.Join(cacheDir, "iso"),
		OcCache:      filepath.Join(cacheDir, "oc"),
		ImageCache:   filepath.Join(cacheDir, "images"),
	}
}
