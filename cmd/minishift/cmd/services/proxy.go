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

package services

import (
	"github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/pkg/minishift/network/proxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	proxyServerPortFlag = "port"
	proxyUpstreamFlag   = "upstream"
)

var (
	proxyServerPortFromFlag   int
	proxyUpstreamAddrFromFlag string

	servicesProxyCmd = &cobra.Command{
		Use:    "proxy",
		Short:  "Starts a proxy server on host",
		Long:   `Starts a proxy server on host`,
		Run:    runProxy,
		Hidden: true,
	}
)

func init() {
	servicesProxyCmd.Flags().IntVarP(&proxyServerPortFromFlag, proxyServerPortFlag, "p", 3128, "The server port.")
	servicesProxyCmd.Flags().StringVarP(&proxyUpstreamAddrFromFlag, proxyUpstreamFlag, "u", "", "The upstream proxy address.")
	ServicesCmd.AddCommand(servicesProxyCmd)
}

func runProxy(cmd *cobra.Command, args []string) {
	proxyPort := viper.GetInt(config.ServicesProxyPort.Name)
	if proxyPort == 0 {
		proxyPort = proxyServerPortFromFlag
	}

	proxyUpstreamAddr := viper.GetString(config.UpstreamProxy.Name)
	if proxyUpstreamAddr == "" {
		proxyUpstreamAddr = proxyUpstreamAddrFromFlag
	}

	proxy.StartProxy(proxyPort, proxyUpstreamAddr)
}
