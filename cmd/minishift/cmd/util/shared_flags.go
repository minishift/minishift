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

package util

import (
	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/pkg/minishift/constants"
	flag "github.com/spf13/pflag"
	"strings"
)

type stringValue string
type stringSliceValue struct {
	value   *[]string
	changed bool
}
type boolValue bool

var HttpProxyFlag = &flag.Flag{
	Name:      configCmd.HttpProxy.Name,
	Shorthand: "",
	Usage:     "HTTP proxy in the format http://<username>:<password>@<proxy_host>:<proxy_port>. Overrides potential HTTP_PROXY setting in the environment.",
	Value:     NewStringValue("", new(string)),
	DefValue:  "",
}

var HttpsProxyFlag = &flag.Flag{
	Name:      configCmd.HttpsProxy.Name,
	Shorthand: "",
	Usage:     "HTTPS proxy in the format https://<username>:<password>@<proxy_host>:<proxy_port>. Overrides potential HTTPS_PROXY setting in the environment.",
	Value:     NewStringValue("", new(string)),
	DefValue:  "",
}

var AddOnEnvFlag = &flag.Flag{
	Name:      configCmd.AddonEnv.Name,
	Shorthand: "a",
	Usage:     "Specify key-value pairs to be added to the add-on interpolation context.",
	Value:     NewStringSliceValue([]string{}, &[]string{}),
}

func NewStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}
func (s *stringValue) Type() string {
	return "string"
}

func (s *stringValue) String() string {
	return string(*s)
}

func NewStringSliceValue(val []string, p *[]string) *stringSliceValue {
	ssv := new(stringSliceValue)
	ssv.value = p
	*ssv.value = val
	return ssv
}

func (s *stringSliceValue) Set(val string) error {
	if !s.changed {
		*s.value = []string{val}
	} else {
		*s.value = append(*s.value, val)
	}
	s.changed = true
	return nil
}

func (s *stringSliceValue) Type() string {
	return "stringSlice"
}

func (s *stringSliceValue) String() string {
	str, _ := writeSliceAsString(*s.value)
	return str
}

func writeSliceAsString(values []string) (string, error) {
	singleVal := ""
	for _, val := range values {
		singleVal = singleVal + " " + val
	}

	return strings.TrimPrefix(singleVal, " "), nil
}

// initClusterUpFlags creates the CLI flags which needs to be passed on to 'oc cluster up'
func InitClusterUpFlags(commandName string) *flag.FlagSet {
	clusterUpFlagSet := flag.NewFlagSet(commandName, flag.ContinueOnError)

	clusterUpFlagSet.Bool(configCmd.SkipRegistryCheck.Name, false, "Skip the Docker daemon registry check.")
	clusterUpFlagSet.String(configCmd.PublicHostname.Name, "", "Public hostname of the OpenShift cluster.")
	clusterUpFlagSet.String(configCmd.RoutingSuffix.Name, "", "Default suffix for the server routes.")
	clusterUpFlagSet.Int(configCmd.ServerLogLevel.Name, 0, "Log level for the OpenShift server.")
	clusterUpFlagSet.String(configCmd.NoProxyList.Name, "", "List of hosts or subnets for which no proxy should be used.")
	clusterUpFlagSet.String(configCmd.ImageName.Name, "", "Specify the images to use for OpenShift")
	clusterUpFlagSet.AddFlag(HttpProxyFlag)
	clusterUpFlagSet.AddFlag(HttpsProxyFlag)
	// This is hidden because we don't want our users to use this flag
	// we are setting it to openshift/origin-${component}:<user_provided_version> as default
	// It is used for testing purpose from CDK/minishift QE
	clusterUpFlagSet.MarkHidden(configCmd.ImageName.Name)

	if enableExperimental, _ := GetBoolEnv(constants.MinishiftEnableExperimental); enableExperimental {
		clusterUpFlagSet.String(configCmd.ExtraClusterUpFlags.Name, "", "Specify optional flags for use with 'cluster up' (unsupported)")
		clusterUpFlagSet.Bool(configCmd.WriteConfig.Name, false, "Write the configuration files into host config dir")
	}

	return clusterUpFlagSet
}
