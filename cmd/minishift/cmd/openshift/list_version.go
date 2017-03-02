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
	"os"

	"github.com/minishift/minishift/pkg/minikube/openshiftversions"
	"github.com/spf13/cobra"
)

// getVersionsCmd represents the ip command
var getVersionsCmd = &cobra.Command{
	Use:   "list-version",
	Short: "Gets the list of OpenShift versions available for Minishift.",
	Long:  `Gets the list of OpenShift versions available for Minishift.`,
	Run: func(cmd *cobra.Command, args []string) {
		openshiftversions.PrintOpenShiftVersionsFromGitHub(os.Stdout)
	},
}

func init() {
	OpenShiftConfigCmd.AddCommand(getVersionsCmd)
}
