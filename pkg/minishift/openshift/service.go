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
	"path/filepath"
	"strconv"
	"strings"

	"encoding/json"

	"github.com/minishift/minishift/pkg/minikube/constants"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util"
	"github.com/pkg/errors"
)

var (
	systemKubeConfigPath string
	runner               util.Runner = &util.RealRunner{}
)

func init() {
	systemKubeConfigPath = filepath.Join(constants.Minipath, "machines", constants.MachineName+"_kubeconfig")
}

type Service struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Spec struct {
			Ports []struct {
				NodePort int `json:"nodePort"`
			} `json:"ports"`
		} `json:"spec"`
	} `json:"items"`
}

type Route struct {
	Items []struct {
		Spec struct {
			AlternateBackends []struct {
				Name   string `json:"name"`
				Weight int    `json:"weight"`
			} `json:"alternateBackends"`
			Host string `json:"host"`
			To   struct {
				Name   string `json:"name"`
				Weight int    `json:"weight"`
			} `json:"to"`
		} `json:"spec"`
	} `json:"items"`
}

type ServiceWeight struct {
	Name   string
	Weight string
}

type ServiceSpec struct {
	Namespace string
	Name      string
	URL       []string
	NodePort  string
	Weight    []string
}

const (
	ProjectsCustomCol = "-o=custom-columns=NAME:.metadata.name"
)

// GetServiceSpecs takes Namespace string and return route/nodeport/name/weight in a ServiceSpec structure
func GetServiceSpecs(serviceNamespace string) ([]ServiceSpec, error) {
	var serviceSpecs []ServiceSpec

	if serviceNamespace != "" && !isProjectExists(serviceNamespace) {
		return serviceSpecs, errors.New(fmt.Sprintf("Namespace %s doesn't exits", serviceNamespace))
	}

	namespaces, err := getValidNamespaces(serviceNamespace)
	if err != nil {
		return serviceSpecs, err
	}

	// iterate over namespaces, get command output, format route and nodePort
	for _, namespace := range namespaces {
		serviceNodePort, err := getService(namespace)
		if err != nil {
			return serviceSpecs, err
		}
		routeSpec, err := getRouteSpec(namespace)
		if err != nil {
			return serviceSpecs, err
		}
		if serviceNodePort != nil {
			serviceSpecs = append(serviceSpecs, filterAndUpdateServiceSpecs(routeSpec, namespace, serviceNodePort)...)
		}
	}
	if len(serviceSpecs) == 0 {
		return serviceSpecs, errors.New(fmt.Sprintf("No services defined in namespace '%s'", serviceNamespace))
	}

	return serviceSpecs, nil
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

// getRouteSpec take valid namespace and return map which have route as key and
// value be ServiceWeight struct
func getRouteSpec(namespace string) (map[string][]ServiceWeight, error) {
	routeSpec := make(map[string][]ServiceWeight)
	cmdArgText := fmt.Sprintf("get route -o json --config=%s -n %s", systemKubeConfigPath, namespace)
	tokens := strings.Split(cmdArgText, " ")
	cmdName := instanceState.InstanceConfig.OcPath
	cmdOut, err := runner.Output(cmdName, tokens...)
	if err != nil {
		return nil, err
	}

	if strings.Contains(string(cmdOut), "No resources found") {
		return nil, nil
	}

	var data Route
	err = json.Unmarshal(cmdOut, &data)
	if err != nil {
		return nil, err
	}

	for _, item := range data.Items {
		totalWeight := item.Spec.To.Weight
		for _, alternateBackend := range item.Spec.AlternateBackends {
			totalWeight += alternateBackend.Weight
		}
		routeSpec[item.Spec.Host] = []ServiceWeight{{Name: item.Spec.To.Name, Weight: calculateWeight(float64(item.Spec.To.Weight), float64(totalWeight))}}
		for _, alternateBackend := range item.Spec.AlternateBackends {
			routeSpec[item.Spec.Host] = append(routeSpec[item.Spec.Host], ServiceWeight{Name: alternateBackend.Name,
				Weight: calculateWeight(float64(alternateBackend.Weight), float64(totalWeight))})
		}
	}
	return routeSpec, nil
}

func calculateWeight(a float64, b float64) string {
	weightPercentage := int(a / b * 100)
	if weightPercentage != 100 {
		return strconv.Itoa(weightPercentage) + "%"
	}
	return ""
}

func filterAndUpdateServiceSpecs(routeSpec map[string][]ServiceWeight, namespace string, serviceNodePort map[string]string) []ServiceSpec {
	var serviceSpecs []ServiceSpec
	serviceSpecMap := make(map[string]*ServiceSpec)

	for routeURL, v := range routeSpec {
		for _, service := range v {
			// split content on colon to separate NAME, HOST and Weight
			if _, ok := serviceNodePort[service.Name]; ok {
				if _, ok := serviceSpecMap[service.Name]; !ok {
					serviceSpecMap[service.Name] = &ServiceSpec{Namespace: namespace,
						Name:     service.Name,
						URL:      []string{fmt.Sprintf("http://%s", routeURL)},
						NodePort: serviceNodePort[service.Name],
						Weight:   []string{service.Weight},
					}
				} else {
					serviceSpecMap[service.Name].URL = append(serviceSpecMap[service.Name].URL, fmt.Sprintf("http://%s", routeURL))
					serviceSpecMap[service.Name].Weight = append(serviceSpecMap[service.Name].Weight, service.Weight)
				}
				delete(serviceNodePort, service.Name)
			} else {
				if _, ok := serviceSpecMap[service.Name]; !ok {
					serviceSpecMap[service.Name] = &ServiceSpec{Namespace: namespace,
						Name:     service.Name,
						URL:      []string{fmt.Sprintf("http://%s", routeURL)},
						NodePort: "",
						Weight:   []string{service.Weight},
					}
				} else {
					serviceSpecMap[service.Name].URL = append(serviceSpecMap[service.Name].URL, fmt.Sprintf("http://%s", routeURL))
					serviceSpecMap[service.Name].Weight = append(serviceSpecMap[service.Name].Weight, service.Weight)
				}
			}
		}
	}

	for k, v := range serviceNodePort {
		serviceSpecs = append(serviceSpecs, ServiceSpec{Namespace: namespace, Name: k, URL: nil, NodePort: v, Weight: nil})
	}
	for _, v := range serviceSpecMap {
		serviceSpecs = append(serviceSpecs, *v)
	}

	return serviceSpecs
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

// Get service provide service name and node-port as map data structure
func getService(namespace string) (map[string]string, error) {
	serviceNodePort := make(map[string]string)
	cmdArgText := fmt.Sprintf("get svc -o json --config=%s -n %s", systemKubeConfigPath, namespace)
	tokens := strings.Split(cmdArgText, " ")
	cmdName := instanceState.InstanceConfig.OcPath
	cmdOut, err := runner.Output(cmdName, tokens...)
	if err != nil {
		return nil, err
	}

	if strings.Contains(string(cmdOut), "No resources found") {
		return nil, nil
	}

	var data Service
	err = json.Unmarshal(cmdOut, &data)
	if err != nil {
		return nil, err
	}

	for _, item := range data.Items {
		for _, port := range item.Spec.Ports {
			nodePort := ""
			if port.NodePort != 0 {
				nodePort = fmt.Sprintf("%d", port.NodePort)
			}
			serviceNodePort[item.Metadata.Name] = nodePort
		}
	}
	return serviceNodePort, nil
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

	contents := strings.Split(string(cmdOut), "\n")
	return emptyFilter(contents[1:]), nil
}
