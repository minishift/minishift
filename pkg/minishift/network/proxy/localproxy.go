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
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	goos "os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/elazarl/goproxy"
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/minishift/config"
	minishiftTLS "github.com/minishift/minishift/pkg/minishift/tls"
	"github.com/minishift/minishift/pkg/util/os"
	"github.com/minishift/minishift/pkg/util/os/process"
)

const (
	proxyAuthHeader = "Proxy-Authorization"
)

func StartProxy(proxyPort int, proxyUpstreamAddr string, reEncrypt bool) {
	// set custom CA
	setCA(minishiftTLS.CACert, minishiftTLS.CAKey)

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.KeepProxyHeaders = true
	if reEncrypt {
		proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	}

	bindAddr := fmt.Sprintf("0.0.0.0:%d", proxyPort)
	chain := "" // for logging

	if proxyUpstreamAddr != "" {

		credentials := ""
		if strings.Index(proxyUpstreamAddr, "@") > -1 {
			u, _ := url.Parse(proxyUpstreamAddr)
			credentials = u.User.String()
			proxyUpstreamAddr = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		}

		chain = fmt.Sprintf(" -> %s", proxyUpstreamAddr)
		proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(proxyUpstreamAddr)
		}

		connectReqHandler := func(req *http.Request) {
			if credentials != "" {
				SetBasicAuth(credentials, req)
			}
		}

		proxy.ConnectDial = proxy.NewConnectDialToProxyWithHandler(proxyUpstreamAddr, connectReqHandler)
		proxy.OnRequest().Do(goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			if credentials != "" {
				SetBasicAuth(credentials, req)
			}
			return req, nil
		}))
	}

	log.Println(fmt.Sprintf("Serving as HTTP proxy on %s%s", bindAddr, chain))
	log.Fatal(http.ListenAndServe(bindAddr, proxy))
}

func SetBasicAuth(credentials string, req *http.Request) {
	authHeader := fmt.Sprintf("Basic %s", basicAuth(credentials))
	req.Header.Add(proxyAuthHeader, authHeader)
}

func basicAuth(credentials string) string {
	return base64.StdEncoding.EncodeToString([]byte(credentials))
}

func setCA(caCert, caKey []byte) error {
	goproxyCa, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return err
	}
	if goproxyCa.Leaf, err = x509.ParseCertificate(goproxyCa.Certificate[0]); err != nil {
		return err
	}
	goproxy.GoproxyCa = goproxyCa
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&goproxyCa)}
	return nil
}

func EnsureProxyDaemonRunning() error {
	if isRunning() {
		if glog.V(2) {
			fmt.Println(fmt.Sprintf("proxy running with pid %d", config.AllInstancesConfig.ProxyPID))
		}
		return nil
	}

	proxyCmd, err := createProxyCommand()
	if err != nil {
		return err
	}

	err = proxyCmd.Start()
	if err != nil {
		return err
	}

	config.AllInstancesConfig.ProxyPID = proxyCmd.Process.Pid
	config.AllInstancesConfig.Write()
	return nil
}

func isRunning() bool {
	if config.AllInstancesConfig.ProxyPID <= 0 {
		return false
	}

	process, err := goos.FindProcess(config.AllInstancesConfig.ProxyPID)
	if err != nil {
		return false
	}

	// for Windows FindProcess is enough
	if runtime.GOOS == "windows" {
		return true
	}

	// for non Windows we need to send a signal to get more information
	err = process.Signal(syscall.Signal(0))
	if err == nil {
		return true
	} else {
		return false
	}
}

func createProxyCommand() (*exec.Cmd, error) {
	cmd, err := os.CurrentExecutable()
	if err != nil {
		return nil, err
	}

	args := []string{
		"services",
		"proxy"}
	exportCmd := exec.Command(cmd, args...)
	// don't inherit any file handles
	exportCmd.Stderr = nil
	exportCmd.Stdin = nil
	exportCmd.Stdout = nil
	exportCmd.SysProcAttr = process.SysProcForBackgroundProcess()
	exportCmd.Env = process.EnvForBackgroundProcess()

	return exportCmd, nil
}
