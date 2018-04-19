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

package cmd

import (
	"fmt"

	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/pkg/minishift/oc"
	"github.com/minishift/minishift/pkg/util"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// preflightChecksForArtifacts is executed once artifacts are cached.
func preflightChecksForArtifacts() {
	preflightCheckSucceedsOrFails(
		configCmd.SkipCheckClusterUpFlag.Name,
		checkOcFlag,
		"Checking if provided oc flags are supported",
		configCmd.WarnCheckClusterUpFlag.Name,
		"Provided oc flag not supported")
}

// checkOcFlag checks if provided oc flags are supported
func checkOcFlag() bool {
	clusterUpParams := determineInitialClusterupParameters()
	for _, key := range clusterUpParams {

		// no need to check for extra-clusterup-flags as it is a minishift specific flag
		if key == configCmd.ExtraClusterUpFlags.Name {
			continue
		}
		if !oc.SupportFlag(key, ocPath, &util.RealRunner{}) {
			fmt.Printf("Flag '%s' is not supported for oc version %s. Use 'openshift-version' flag to select a different version of OpenShift.\n", key, viper.GetString(configCmd.OpenshiftVersion.Name))
			return false
		}
	}
	return true
}

// determineIntialClusterupParameters return the list of used oc cluster up parameters during start
func determineInitialClusterupParameters() []string {
	var clusterUpParams []string
	clusterUpFlagSet.VisitAll(func(flag *flag.Flag) {
		if viper.IsSet(flag.Name) {
			key := flag.Name
			_, exists := minishiftToClusterUp[key]
			if exists {
				key = minishiftToClusterUp[key]
			}
			clusterUpParams = append(clusterUpParams, key)
		}
	})

	return clusterUpParams
}
