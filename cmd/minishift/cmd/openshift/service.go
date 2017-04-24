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

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/minishift/minishift/pkg/minishift/openshift"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

var (
	namespace string
	urlMode   bool
	https     bool
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service [flags] SERVICE",
	Short: "Opens the URL for the specified service in the browser or prints it to the console",
	Long:  `Opens the URL for the specified service and namespace in the browser or prints it to the console. If no namespace is provided, 'default' is assumed.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			atexit.ExitWithMessage(1, "You must specify the name of the service.")
		}

		service := args[0]

		url, err := openshift.GetServiceURL(service, namespace, https)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		if urlMode {
			fmt.Fprintln(os.Stdout, url)
		} else {
			fmt.Fprintln(os.Stdout, "Opening the service "+namespace+"/"+service+" in the default browser...")
			browser.OpenURL(url)
		}
	},
}

func init() {
	serviceCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "The namespace of the service.")
	serviceCmd.Flags().BoolVar(&urlMode, "url", false, "Access the service in the command-line console instead of the default browser.")
	serviceCmd.Flags().BoolVar(&https, "https", false, "Access the service with HTTPS instead of HTTP.")
	OpenShiftCmd.AddCommand(serviceCmd)
}
