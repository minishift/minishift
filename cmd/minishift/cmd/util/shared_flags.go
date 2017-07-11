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
	flag "github.com/spf13/pflag"
	"strings"
)

const (
	HttpProxy  = "http-proxy"
	HttpsProxy = "https-proxy"
	AddOnEnv   = "addon-env"
)

type stringValue string
type stringSliceValue struct {
	value   *[]string
	changed bool
}

var HttpProxyFlag = &flag.Flag{
	Name:      HttpProxy,
	Shorthand: "",
	Usage:     "HTTP proxy in the format http://<username>:<password>@<proxy_host>:<proxy_port>. Overrides potential HTTP_PROXY setting in the environment.",
	Value:     newStringValue("", new(string)),
	DefValue:  "",
}

var HttpsProxyFlag = &flag.Flag{
	Name:      HttpsProxy,
	Shorthand: "",
	Usage:     "HTTPS proxy in the format https://<username>:<password>@<proxy_host>:<proxy_port>. Overrides potential HTTPS_PROXY setting in the environment.",
	Value:     newStringValue("", new(string)),
	DefValue:  "",
}

var AddOnEnvFlag = &flag.Flag{
	Name:      AddOnEnv,
	Shorthand: "a",
	Usage:     "Specify key-value pairs to be added to the add-on interpolation context.",
	Value:     newStringSliceValue([]string{}, &[]string{}),
}

func newStringValue(val string, p *string) *stringValue {
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

func newStringSliceValue(val []string, p *[]string) *stringSliceValue {
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
