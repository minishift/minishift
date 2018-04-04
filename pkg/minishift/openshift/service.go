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
	"strconv"
	"strings"

	"encoding/json"

	"github.com/minishift/minishift/pkg/minikube/constants"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util"
	"github.com/pkg/errors"
)

var (
	runner util.Runner = &util.RealRunner{}
)

const (
	http  = "http://%s"
	https = "https://%s"
)

type ServiceSpec struct {
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

type RouteSpec struct {
	Items []struct {
		Spec struct {
			AlternateBackends []struct {
				Name   string `json:"name"`
				Weight int    `json:"weight"`
			} `json:"alternateBackends"`
			Host string `json:"host"`
			Tls  struct {
				Termination string `json:"termination"`
			} `json:"tls"`
			To struct {
				Name   string `json:"name"`
				Weight int    `json:"weight"`
			} `json:"to"`
		} `json:"spec"`
	} `json:"items"`
}

type Route struct {
	Name   string
	Weight string
	Tls    string
}

type Service struct {
	Namespace string
	Name      string
	URL       []string
	NodePort  string
	Weight    []string
}

type Services struct {
	Service map[string]*Service
}

const (
	ProjectsCustomCol = "-o=custom-columns=NAME:.metadata.name"
)

// GetServices takes Namespace string and return route/nodeport/name/weight in a Service structure
func GetServices(serviceNamespace string) ([]Service, error) {
	var services []Service

	if serviceNamespace != "" && !isProjectExists(serviceNamespace) {
		return services, errors.New(fmt.Sprintf("Namespace '%s' doesn't exist", serviceNamespace))
	}

	namespaces, err := getValidNamespaces(serviceNamespace)
	if err != nil {
		return services, err
	}

	// iterate over namespaces, get command output, format route and nodePort
	for _, namespace := range namespaces {
		serviceNodePort, err := getNodePorts(namespace)
		if err != nil {
			return services, err
		}
		route, err := getRoutes(namespace)
		if err != nil {
			return services, err
		}
		if serviceNodePort != nil {
			services = append(services, filterAndUpdateServices(route, namespace, serviceNodePort)...)
		}
	}
	if len(services) == 0 {
		return services, errors.New(fmt.Sprintf("No services defined in namespace '%s'", serviceNamespace))
	}

	return services, nil
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
// value be Route struct
func getRoutes(namespace string) (map[string][]Route, error) {
	routes := make(map[string][]Route)
	cmdArgText := fmt.Sprintf("get route -o json --config=%s -n %s", constants.KubeConfigPath, namespace)
	tokens := strings.Split(cmdArgText, " ")
	cmdName := instanceState.InstanceConfig.OcPath
	cmdOut, err := runner.Output(cmdName, tokens...)
	if err != nil {
		return nil, err
	}

	if strings.Contains(string(cmdOut), "No resources found") {
		return nil, nil
	}

	var data RouteSpec
	err = json.Unmarshal(cmdOut, &data)
	if err != nil {
		return nil, err
	}

	for _, item := range data.Items {
		totalWeight := item.Spec.To.Weight
		for _, alternateBackend := range item.Spec.AlternateBackends {
			totalWeight += alternateBackend.Weight
		}
		routes[item.Spec.Host] = []Route{{Name: item.Spec.To.Name, Weight: calculateWeight(float64(item.Spec.To.Weight), float64(totalWeight)), Tls: item.Spec.Tls.Termination}}
		for _, alternateBackend := range item.Spec.AlternateBackends {
			routes[item.Spec.Host] = append(routes[item.Spec.Host], Route{Name: alternateBackend.Name,
				Weight: calculateWeight(float64(alternateBackend.Weight), float64(totalWeight)), Tls: item.Spec.Tls.Termination})
		}
	}
	return routes, nil
}

func calculateWeight(a float64, b float64) string {
	weightPercentage := int(a / b * 100)
	if weightPercentage != 100 {
		return strconv.Itoa(weightPercentage) + "%"
	}
	return ""
}

func hasNodePort(serviceNodePort map[string]string, name string) bool {
	_, ok := serviceNodePort[name]
	return ok
}

func createService(route Route, routeURL, namespace, nodePort string) *Service {
	s := &Service{Name: route.Name,
		Namespace: namespace,
		Weight:    []string{route.Weight},
		NodePort:  nodePort,
	}

	s.URL = append(s.URL, genUrl(route, routeURL))
	return s
}

func genUrl(route Route, url string) string {
	if route.Tls != "" {
		return fmt.Sprintf(https, url)
	}
	return fmt.Sprintf(http, url)
}

func NewServices() *Services {
	return &Services{Service: make(map[string]*Service)}
}

func (s *Services) Add(route Route, routeURL, namespace, nodePort string) {
	if s.HasService(route.Name) {
		s.Service[route.Name].URL = append(s.Service[route.Name].URL, genUrl(route, routeURL))
		s.Service[route.Name].Weight = append(s.Service[route.Name].Weight, route.Weight)
	} else {
		s.Service[route.Name] = createService(route, routeURL, namespace, nodePort)
	}
}

func (s *Services) HasService(name string) bool {
	_, ok := s.Service[name]
	return ok
}

func filterAndUpdateServices(routes map[string][]Route, namespace string, serviceNodePort map[string]string) []Service {
	var serviceList []Service
	services := NewServices()
	for routeURL, v := range routes {
		for _, service := range v {
			if hasNodePort(serviceNodePort, service.Name) {
				services.Add(service, routeURL, namespace, serviceNodePort[service.Name])
				delete(serviceNodePort, service.Name)
			} else {
				services.Add(service, routeURL, namespace, "")
			}
		}
	}

	for k, v := range serviceNodePort {
		serviceList = append(serviceList, Service{Namespace: namespace, Name: k, URL: nil, NodePort: v, Weight: nil})
	}
	for _, v := range services.Service {
		serviceList = append(serviceList, *v)
	}

	return serviceList
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

// getNodePorts  provide service name and node-port as map data structure
func getNodePorts(namespace string) (map[string]string, error) {
	serviceNodePorts := make(map[string]string)
	cmdArgText := fmt.Sprintf("get svc -o json --config=%s -n %s", constants.KubeConfigPath, namespace)
	tokens := strings.Split(cmdArgText, " ")
	cmdName := instanceState.InstanceConfig.OcPath
	cmdOut, err := runner.Output(cmdName, tokens...)
	if err != nil {
		return nil, err
	}

	if strings.Contains(string(cmdOut), "No resources found") {
		return nil, nil
	}

	var data ServiceSpec
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
			serviceNodePorts[item.Metadata.Name] = nodePort
		}
	}
	return serviceNodePorts, nil
}

// Get all projects a user belongs to
func getProjects() ([]string, error) {
	cmdArgText := fmt.Sprintf("get projects --config=%s %s", constants.KubeConfigPath, ProjectsCustomCol)
	tokens := strings.Split(cmdArgText, " ")
	cmdName := instanceState.InstanceConfig.OcPath
	cmdOut, err := runner.Output(cmdName, tokens...)
	if err != nil {
		return []string{}, err
	}

	contents := strings.Split(string(cmdOut), "\n")
	return emptyFilter(contents[1:]), nil
}
