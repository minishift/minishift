// +build integration

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

package proxy

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/elazarl/goproxy"
	minishiftUtil "github.com/minishift/minishift/pkg/minishift/util"
)

var (
	proxyListener net.Listener
	proxyLog      *bytes.Buffer
)

const DEFAULT_PROXY_PORT string = "8181"

func GetLog() string {
	var log string
	if proxyLog != nil {
		log = proxyLog.String()
	}

	return log
}

func ResetLog(verbose bool) {
	if GetLog() != "" {
		if verbose == true {
			fmt.Printf("beginning of proxy log >>>\n%s<<< end of proxy log\n", GetLog())
		}
		proxyLog.Reset()
	}

	return
}

func startProxy(proxyPort string) error {
	var err error
	proxyListener, err = net.Listen("tcp", ":"+proxyPort)
	if err != nil {
		return err
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxyLog = bytes.NewBufferString("")
	proxy.Logger = log.New(proxyLog, "", log.Ldate)

	http.Serve(proxyListener, proxy)
	fmt.Println("Proxy has stopped.")

	return nil
}

func getPort() string {
	if value, present := os.LookupEnv("INTEGRATION_PROXY_CUSTOM_PORT"); present == true {
		return value
	}

	return DEFAULT_PROXY_PORT
}

func getIP() (string, error) {
	if value, present := os.LookupEnv("INTEGRATION_PROXY_CUSTOM_IP"); present == true {
		return value, nil
	}

	ips := minishiftUtil.HostIPs()
	if ips == nil {
		return "", errors.New(`No IP found. This might be an error in automated detection of available network devices.
							   You can use INTEGRATION_PROXY_CUSTOM_IP to set IP manually.`)
	}

	ip := strings.Split(ips[0], "/")[0]
	return ip, nil
}

func SetProxy() error {
	port := getPort()
	ip, err := getIP()
	if err != nil {
		return err
	}

	go startProxy(port)

	address := fmt.Sprintf("http://%v:%v", ip, port)
	os.Setenv("MINISHIFT_HTTP_PROXY", address)
	fmt.Printf("  INFO: this feature is now using MINISHIFT_HTTP_PROXY=%s\n", os.Getenv("MINISHIFT_HTTP_PROXY"))

	return nil
}

func UnsetProxy() error {
	os.Unsetenv("MINISHIFT_HTTP_PROXY")
	if proxyListener != nil {
		proxyListener.Close()
	}

	return nil
}
