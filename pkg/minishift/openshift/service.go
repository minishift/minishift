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

package openshift

import (
	"fmt"
	"github.com/minishift/minishift/pkg/minikube/constants"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util"
	"github.com/pkg/errors"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	systemKubeConfigPath string
	runner               util.Runner = &util.RealRunner{}
)

func init() {
	systemKubeConfigPath = filepath.Join(constants.Minipath, "machines", constants.MachineName+"_kubeconfig")
}

type ServiceURL struct {
	Namespace string
	Name      string
	URL       string
}

const (
	URLCustomCol      = "-o=custom-columns=URL:.spec.host"
	URLsCustomCol     = "-o=custom-columns=NAME:.metadata.name,HOST:.spec.host"
	ProjectsCustomCol = "-o=custom-columns=NAME:.metadata.name"
)

// Get the route for service
func GetServiceURL(service, namespace string, https bool) (string, error) {
	urlScheme := "http://"
	if https {
		urlScheme = "https://"
	}

	if !isProjectExists(namespace) {
		return "", errors.New(fmt.Sprintf("Namespace %s doesn't exits", namespace))
	}

	cmdArgText := fmt.Sprintf("get route/%s -n %s --config=%s %s", service, namespace, systemKubeConfigPath, URLCustomCol)
	tokens := strings.Split(cmdArgText, " ")
	cmdName := instanceState.InstanceConfig.OcPath
	cmdOut, err := runner.Output(cmdName, tokens...)
	if err != nil {
		return "", errors.New(fmt.Sprintf("No service with name '%s' defined in namespace '%s'", service, namespace))
	}

	url := strings.Split(byteArrayToString(cmdOut), "\n")[1] // second element contain actual URL content
	return urlScheme + url, nil
}

// Get the available routes to user
func GetServiceURLs(serviceListNamespace string) ([]ServiceURL, error) {
	var serviceURLs []ServiceURL

	if serviceListNamespace != "" && !isProjectExists(serviceListNamespace) {
		return serviceURLs, errors.New(fmt.Sprintf("Namespace %s doesn't exits", serviceListNamespace))
	}

	namespaces, err := getValidNamespaces(serviceListNamespace)
	if err != nil {
		return serviceURLs, err
	}

	// iterate over namespaces, get command output, format route in ServiceURL
	for _, namespace := range namespaces {
		outputData, err := getServiceURLsOutput(namespace)
		if err != nil {
			return serviceURLs, err
		}
		if !strings.Contains(outputData, "No resources found") {
			serviceURLs = filterAndUpdateServiceURLS(outputData, namespace)
		}
	}
	if len(serviceURLs) == 0 {
		return serviceURLs, errors.New(fmt.Sprintf("No services defined in namespace '%s'", serviceListNamespace))
	}

	return serviceURLs, nil
}

// Check whether project exists or not
func isProjectExists(projectName string) bool {
	projects, err := getProjects()
	if err != nil {
		return false
	}

	for _, name := range projects {
		if name == projectName {
			return true
		}
	}

	return false
}

func getValidNamespaces(serviceListNamespace string) ([]string, error) {
	var (
		namespaces []string
		err        error
	)

	// If namespace is default then consider all namespaces user belongs to
	if serviceListNamespace == "" {
		namespaces, err = getProjects()
		if err != nil {
			return namespaces, errors.Wrap(err, "Error getting valid namespaces user belongs to.")
		}
	} else {
		namespaces = append(namespaces, serviceListNamespace)
	}

	return namespaces, nil
}

func getServiceURLsOutput(namespace string) (string, error) {
	cmdArgText := fmt.Sprintf("get route -n %s --config=%s %s", namespace, systemKubeConfigPath, URLsCustomCol)
	tokens := strings.Split(cmdArgText, " ")
	cmdName := instanceState.InstanceConfig.OcPath
	cmdOut, err := runner.Output(cmdName, tokens...)
	if err != nil {
		return "", err
	}

	return byteArrayToString(cmdOut), nil
}

func filterAndUpdateServiceURLS(data, namespace string) []ServiceURL {
	var serviceURLs []ServiceURL

	re_whtsp_inside := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	data = re_whtsp_inside.ReplaceAllString(data, " ") // replace all extra whitespaces
	contents := strings.Split(data, "\n")              // split on new lines
	contents = emptyFilter(contents[1:])               // remove the header "NAME HOST" and empty elements

	for _, content := range contents {
		// split content on white space to separate NAME and HOST
		data := strings.Split(content, " ")
		serviceURLs = append(serviceURLs, ServiceURL{Namespace: namespace, Name: data[0], URL: "http://" + data[1]})
	}

	return serviceURLs
}

// Discard empty elements
func emptyFilter(data []string) []string {
	var res []string

	for _, ele := range data {
		if ele != "" {
			res = append(res, ele)
		}
	}

	return res
}

// Get all projects a user belongs to
func getProjects() ([]string, error) {
	cmdArgText := fmt.Sprintf("get projects --config=%s %s", systemKubeConfigPath, ProjectsCustomCol)
	tokens := strings.Split(cmdArgText, " ")
	cmdName := instanceState.InstanceConfig.OcPath
	cmdOut, err := runner.Output(cmdName, tokens...)
	if err != nil {
		return []string{}, err
	}

	contents := strings.Split(byteArrayToString(cmdOut), "\n")
	return emptyFilter(contents[1:]), nil
}

// Convert byte array to string
func byteArrayToString(data []byte) string {
	return string(data)
}
