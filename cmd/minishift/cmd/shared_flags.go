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

package cmd

import (
	flag "github.com/spf13/pflag"
)

type stringValue string

var httpProxyFlag = &flag.Flag{
	Name:      httpProxy,
	Shorthand: "",
	Usage:     "HTTP proxy used for downloading artefact and configure Docker as well as OpenShift (In the format of http://<username>:<password>@<proxy_host>:<proxy_port>). Overrides a potential HTTP_PROXY setting in the environment.",
	Value:     newStringValue("", new(string)),
	DefValue:  "",
}

var httpsProxyFlag = &flag.Flag{
	Name:      httpsProxy,
	Shorthand: "",
	Usage:     "HTTPS proxy used for downloading artefact and configure Docker as well as OpenShift (In the format of https://<username>:<password>@<proxy_host>:<proxy_port>). Overrides a potential HTTPS_PROXY setting in the environment.",
	Value:     newStringValue("", new(string)),
	DefValue:  "",
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
