/*
Copyright (C) 2016 Red Hat, Inc.

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

package cache

import (
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/github"
	minishiftos "github.com/minishift/minishift/pkg/util/os"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

const OC_CACHE_DIR = "oc"

// Oc is a struct with methods designed for dealing with the oc binary
type Oc struct {
	OpenShiftVersion  string
	MinishiftCacheDir string
}

func (oc *Oc) EnsureIsCached() error {
	if !oc.isCached() {
		err := oc.cacheOc()
		if err != nil {
			return err
		}

	}
	return nil
}

func (oc *Oc) GetCacheFilepath() string {
	return filepath.Join(oc.MinishiftCacheDir, OC_CACHE_DIR, oc.OpenShiftVersion)
}

func (oc *Oc) isCached() bool {
	if _, err := os.Stat(filepath.Join(oc.GetCacheFilepath(), constants.OC_BINARY_NAME)); os.IsNotExist(err) {
		return false
	}
	return true
}

// cacheOc downloads and caches the oc binary into the minishift directory
func (oc *Oc) cacheOc() error {
	if !oc.isCached() {
		if err := github.DownloadOpenShiftReleaseBinary(github.OC, minishiftos.CurrentOS(), oc.OpenShiftVersion, oc.GetCacheFilepath()); err != nil {
			return errors.Wrapf(err, "Error attempting to download and cache %s", github.OC.String())
		}
	}
	return nil
}
