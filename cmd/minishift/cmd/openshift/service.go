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
	"text/template"

	"github.com/docker/machine/libmachine"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

var (
	namespace          string
	serviceURLMode     bool
	serviceURLFormat   string
	serviceURLTemplate *template.Template
	https              bool
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service [flags] SERVICE",
	Short: "Prints the URL for the specified service to the console.",
	Long:  `Prints the URL for the specified service to the console.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		t, err := template.New("serviceURL").Parse(serviceURLFormat)
		if err != nil {
			fmt.Fprintln(os.Stderr, "The URL format specified in the --format option is not valid: \n\n", err)
			atexit.Exit(1)
		}
		serviceURLTemplate = t
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			fmt.Fprintln(os.Stderr, "You must specify the name of the service.")
			atexit.Exit(1)
		}

		service := args[0]

		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		url, err := cluster.GetServiceURL(api, namespace, service, serviceURLTemplate)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			if _, ok := err.(cluster.MissingNodePortError); !ok {
				fmt.Fprintln(os.Stderr, "Verify that Minishift is running and that the correct namespace is specified in the -n option.")
			}
			atexit.Exit(1)
		}

		if https {
			url = strings.Replace(url, "http", "https", 1)
		}
		if serviceURLMode {
			fmt.Fprintln(os.Stdout, url)
		} else {
			fmt.Fprintln(os.Stdout, "Opening the service "+namespace+"/"+service+" in the default browser...")
			browser.OpenURL(url)
		}
	},
}

func init() {
	serviceCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "The namespace of the service.")
	serviceCmd.Flags().BoolVar(&serviceURLMode, "url", false, "Access the service in the command-line console instead of the default browser.")
	serviceCmd.PersistentFlags().StringVar(&serviceURLFormat, "format", "http://{{.IP}}:{{.Port}}", "The URL format of the service.")
	serviceCmd.Flags().BoolVar(&https, "https", false, "Access the service with HTTPS instead of HTTP.")
	OpenShiftConfigCmd.AddCommand(serviceCmd)
}
