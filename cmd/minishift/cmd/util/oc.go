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

package util

import (
	"fmt"
	"path/filepath"

	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/cache"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

// cacheOc ensures that the oc binary matching the requested OpenShift version is cached on the host
func CacheOc(openShiftVersion string) string {
	ocBinary := cache.Oc{
		OpenShiftVersion:  openShiftVersion,
		MinishiftCacheDir: filepath.Join(constants.Minipath, "cache"),
	}
	if err := ocBinary.EnsureIsCached(); err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error starting the cluster: %v", err))
	}

	// Update MACHINE_NAME.json for oc path
	minishiftConfig.InstanceConfig.OcPath = filepath.Join(ocBinary.GetCacheFilepath(), constants.OC_BINARY_NAME)
	if err := minishiftConfig.InstanceConfig.Write(); err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error updating oc path in config of VM: %v", err))
	}

	return minishiftConfig.InstanceConfig.OcPath
}
