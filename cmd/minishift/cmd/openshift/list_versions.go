/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package openshift

import (
	openshiftVersions "github.com/minishift/minishift/pkg/minishift/openshift/version"
	"github.com/minishift/minishift/pkg/version"
	"github.com/spf13/cobra"
	"os"
)

// getVersionsCmd represents the ip command
var getVersionsCmd = &cobra.Command{
	Use:   "list",
	Short: "Gets the list of OpenShift versions that are available for Minishift.",
	Long:  `Gets the list of OpenShift versions that are available for Minishift.`,
	Run: func(cmd *cobra.Command, args []string) {
		openshiftVersions.PrintUpStreamVersions(os.Stdout, version.GetOpenShiftVersion())
	},
}

func init() {
	versionCmd.AddCommand(getVersionsCmd)
}
