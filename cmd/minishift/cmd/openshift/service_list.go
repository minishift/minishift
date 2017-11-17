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

package openshift

import (
	"fmt"
	"os"
	"strings"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/openshift"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var serviceListNamespace string

// serviceListCmd represents the service list command
var serviceListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "Gets the URLs of the services in your local OpenShift cluster.",
	Long:  `Gets the URLs of the services in your local OpenShift cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
		defer api.Close()

		util.ExitIfUndefined(api, constants.MachineName)

		host, err := api.Load(constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		util.ExitIfNotRunning(host.Driver, constants.MachineName)

		ip, err := host.Driver.GetIP()
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP: %s", err.Error()))
		}

		serviceSpecs, err := openshift.GetServiceSpecs(serviceListNamespace)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		var data [][]string
		namespace := make(map[string]bool)
		for _, serviceSpec := range serviceSpecs {
			if _, ok := namespace[serviceSpec.Namespace]; ok {
				serviceSpec.Namespace = ""
			} else {
				namespace[serviceSpec.Namespace] = true
			}
			var urls, weights string
			nodePortURL := serviceSpec.NodePort
			if nodePortURL != "" {
				nodePortURL = fmt.Sprintf("%s:%s", ip, nodePortURL)
			}
			if serviceSpec.URL != nil {
				urls = strings.Join(serviceSpec.URL, "\n")
			}
			if serviceSpec.Weight != nil {
				weights = strings.Join(serviceSpec.Weight, "\n")
			}
			data = append(data, []string{serviceSpec.Namespace, serviceSpec.Name, nodePortURL, urls, weights})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Namespace", "Name", "NodePort", "Route-URL", "Weight"})
		table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
		table.SetCenterSeparator("|")
		table.AppendBulk(data) // Add Bulk Data
		table.Render()
	},
}

func init() {
	serviceListCmd.Flags().StringVarP(&serviceListNamespace, "namespace", "n", "", "The namespace of the services.")
	serviceCmd.AddCommand(serviceListCmd)
}
