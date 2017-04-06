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
	"os"

	"github.com/minishift/minishift/pkg/minishift/openshift"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var serviceListNamespace string

// serviceListCmd represents the service list command
var serviceListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "Gets the URLs of the services in your local cluster.",
	Long:  `Gets the URLs of the services in your local cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		urls, err := openshift.GetServiceURLs(serviceListNamespace)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		var data [][]string
		for _, url := range urls {
			data = append(data, []string{url.Namespace, url.Name, url.URL})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Namsepace", "Name", "URL"})
		table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
		table.SetCenterSeparator("|")
		table.AppendBulk(data) // Add Bulk Data
		table.Render()
	},
}

func init() {
	serviceListCmd.Flags().StringVarP(&serviceListNamespace, "namespace", "n", "default", "The namespace of the services.")
	serviceCmd.AddCommand(serviceListCmd)
}
