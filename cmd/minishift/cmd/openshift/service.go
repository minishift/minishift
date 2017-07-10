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
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/openshift"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var (
	namespace string
	inbrowser bool
	https     bool
	url       bool
	service   string
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service [flags] SERVICE",
	Short: "Opens the URL for the specified service in the browser or prints it to the console.",
	Long:  `Opens the URL for the specified service and namespace in the default browser or prints it to the console. If no namespace is provided, 'default' is assumed.`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		util.ExitIfUndefined(api, constants.MachineName)

		if len(args) == 0 || len(args) > 1 {
			atexit.ExitWithMessage(1, "You must specify the name of the service.")
		}

		host, err := api.Load(constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		util.ExitIfNotRunning(host.Driver, constants.MachineName)

		ip, err := host.Driver.GetIP()
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error getting IP: %s", err.Error()))
		}

		service = args[0]

		serviceSpecs, err := openshift.GetServiceSpecs(namespace)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		if url {
			stdOutURL(serviceSpecs, ip)
		} else if inbrowser {
			openInBrowser(serviceSpecs, ip)
		} else {
			printToStdOut(serviceSpecs, ip)
		}
	},
}

func init() {
	serviceCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "The namespace of the service.")
	serviceCmd.Flags().BoolVar(&inbrowser, "in-browser", false, "Access the service in the default browser.")
	serviceCmd.Flags().BoolVar(&url, "url", false, "Print the service URL to standard output.")
	serviceCmd.Flags().BoolVar(&https, "https", false, "Access the service with HTTPS instead of HTTP.")
	OpenShiftCmd.AddCommand(serviceCmd)
}

func openInBrowser(serviceSpecs []openshift.ServiceSpec, ip string) {
	serviceURL := getServiceURL(serviceSpecs, ip)
	fmt.Fprintln(os.Stdout, "Opening the route/NodePort "+serviceURL+" in the default browser...")
	browser.OpenURL(serviceURL)
}

func stdOutURL(serviceSpecs []openshift.ServiceSpec, ip string) {
	serviceURL := getServiceURL(serviceSpecs, ip)
	fmt.Fprintln(os.Stdout, serviceURL)
}

func getServiceURL(serviceSpecs []openshift.ServiceSpec, ip string) string {
	serviceURL := ""
	namespaceList := isServiceInMultipleNamespace(serviceSpecs, service)
	if len(namespaceList) == 0 {
		atexit.ExitWithMessage(1, fmt.Sprintf("Service %s does not exist", service))
	}
	if len(namespaceList) > 1 {
		namespaces := strings.TrimSpace(strings.Join(namespaceList, ", "))
		atexit.ExitWithMessage(1, fmt.Sprintf("Service %s exists in multiple namespaces (%s), you need to chose a specific namespace using -n <namespace>.", service, namespaces))
	}

	for _, serviceSpec := range serviceSpecs {
		if serviceSpec.Name == service {
			if serviceSpec.URL != nil {
				serviceURL = serviceSpec.URL[0]
				return serviceURL

			} else if serviceSpec.NodePort != "" {
				nodePortURL := fmt.Sprintf("%s:%s", ip, serviceSpec.NodePort)
				urlScheme := "http://"
				if https {
					urlScheme = "https://"
				}
				serviceURL = urlScheme + nodePortURL
				return serviceURL
			} else {
				atexit.ExitWithMessage(1, fmt.Sprintf("Service '%s' in namespace '%s' does not have route associated which can be opened in the browser.", serviceSpec.Name, serviceSpec.Namespace))
			}
		}
	}
	return serviceURL
}

func isServiceInMultipleNamespace(serviceSpecs []openshift.ServiceSpec, service string) []string {
	namespceList := []string{}
	for _, serviceSpec := range serviceSpecs {
		if serviceSpec.Name == service {
			namespceList = append(namespceList, serviceSpec.Namespace)
		}
	}
	return namespceList
}

func printToStdOut(serviceSpecs []openshift.ServiceSpec, ip string) {
	var data [][]string
	var urls, weights string

	for _, serviceSpec := range serviceSpecs {
		if serviceSpec.Name == service {
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
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Namsepace", "Name", "NodePort", "Route-URL", "Weight"})
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()
}
