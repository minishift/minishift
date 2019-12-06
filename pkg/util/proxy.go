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
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"strings"
)

const OpenShiftRegistryIp = "172.30.1.1"

var defaultNoProxies = []string{"localhost", "127.0.0.1", OpenShiftRegistryIp}

// ProxyConfig keeps the proxy configuration for the current environment
type ProxyConfig struct {
	httpProxy  string
	httpsProxy string
	noProxy    []string
}

// NewProxyConfig creates a proxy configuration with the specified parameters. If a empty string is passed
// the corresponding environment variable is checked.
func NewProxyConfig(httpProxy string, httpsProxy string, noProxy string) (*ProxyConfig, error) {
	defaultScheme := "http"

	// We only parse the Proxy Special char if it is coming from the config file
	// or user defined flag. In case of Env variable we just ignore parsing.
	if httpProxy != "" {
		httpProxy = parseProxySpecialChar(httpProxy, defaultScheme)
	}
	if httpsProxy != "" {
		httpsProxy = parseProxySpecialChar(httpsProxy, defaultScheme)
	}

	if httpProxy == "" {
		httpProxy = os.Getenv("http_proxy")
		if httpProxy == "" {
			httpProxy = os.Getenv("HTTP_PROXY")
		}
	}
	err := ValidateProxyURL(httpProxy, defaultScheme)
	if err != nil {
		return nil, err
	}

	if httpsProxy == "" {
		httpsProxy = os.Getenv("https_proxy")
		if httpsProxy == "" {
			httpsProxy = os.Getenv("HTTPS_PROXY")
		}
	}
	err = ValidateProxyURL(httpsProxy, defaultScheme)
	if err != nil {
		return nil, err
	}

	np := []string{}
	np = append(np, defaultNoProxies...)

	if noProxy == "" {
		noProxy = os.Getenv("no_proxy")
		if noProxy == "" {
			noProxy = os.Getenv("NO_PROXY")
		}
	}

	if noProxy != "" {
		np = append(np, strings.Split(noProxy, ",")...)
	}

	config := ProxyConfig{
		httpProxy:  httpProxy,
		httpsProxy: httpsProxy,
		noProxy:    np,
	}

	return &config, nil
}

// ProxyConfig returns a the proxy configuration as a slice, one element for each of the potential settings
// HTTP_PROXY, HTTPS_PROXY and NO_PROXY. If proxies are not enabled an empty slice is returned.
func (p *ProxyConfig) ProxyConfig() []string {
	config := []string{}

	if !p.IsEnabled() {
		return config
	}

	if p.httpProxy != "" {
		config = append(config, fmt.Sprintf("HTTP_PROXY=%s", p.httpProxy))
		config = append(config, fmt.Sprintf("http_proxy=%s", p.httpProxy))
	}

	if p.httpsProxy != "" {
		config = append(config, fmt.Sprintf("HTTPS_PROXY=%s", p.httpsProxy))
		config = append(config, fmt.Sprintf("https_proxy=%s", p.httpsProxy))
	}
	config = append(config, fmt.Sprintf("NO_PROXY=%s", p.NoProxy()))
	config = append(config, fmt.Sprintf("no_proxy=%s", p.NoProxy()))

	return config
}

// Override proxy for local/intermediate proxy
func (p *ProxyConfig) OverrideHttpProxy(proxy string) {
	p.httpProxy = proxy
}

// Override proxy for local/intermediate proxy
func (p *ProxyConfig) OverrideHttpsProxy(proxy string) {
	p.httpsProxy = proxy
}

// HttpProxy returns the configured value for the HTTP proxy. The empty string is returned in case HTTP proxy is not set.
func (p *ProxyConfig) HttpProxy() string {
	return p.httpProxy
}

// HttpsProxy returns the configured value for the HTTPS proxy. The empty string is returned in case HTTPS proxy is not set.
func (p *ProxyConfig) HttpsProxy() string {
	return p.httpsProxy
}

// NoProxy returns a comma separated list of hosts for which proxies should not be applied.
func (p *ProxyConfig) NoProxy() string {
	if p.IsEnabled() {
		return strings.Join(p.noProxy, ",")
	} else {
		return ""
	}
}

// AddNoProxy appends the specified host to the list of no proxied hosts.
func (p *ProxyConfig) AddNoProxy(host string) {
	p.noProxy = append(p.noProxy, host)
}

// Sets the current config as environment variables in the current process.
func (p *ProxyConfig) ApplyToEnvironment() {
	if !p.IsEnabled() {
		return
	}

	if p.httpProxy != "" {
		os.Setenv("HTTP_PROXY", p.httpProxy)
		os.Setenv("http_proxy", p.httpProxy)
	}
	if p.httpsProxy != "" {
		os.Setenv("HTTPS_PROXY", p.httpsProxy)
		os.Setenv("https_proxy", p.httpsProxy)
	}
	os.Setenv("NO_PROXY", p.NoProxy())
	os.Setenv("no_proxy", p.NoProxy())
}

// Enabled returns true if at least one proxy (HTTP or HTTPS) is configured. Returns false otherwise.
func (p *ProxyConfig) IsEnabled() bool {
	return p.httpProxy != "" || p.httpsProxy != ""
}

// ValidateProxyURL validates that the specified proxyURL is valid
func ValidateProxyURL(proxyUrl string, defaultScheme string) error {
	if proxyUrl == "" {
		return nil
	}

	if !strings.HasPrefix(proxyUrl, "http://") &&
		!strings.HasPrefix(proxyUrl, "https://") {
		proxyUrl = fmt.Sprintf("%s://%s", defaultScheme, proxyUrl)
	}

	if !govalidator.IsURL(proxyUrl) {
		return errors.Errorf("Proxy URL '%s' is not valid.", proxyUrl)
	}
	return nil
}

// parseProxySpecialChar parse the URI and convert special char to hex
func parseProxySpecialChar(proxyUrl string, defaultScheme string) string {
	if proxyUrl == "" {
		return ""
	}

	proxyUrl = strings.TrimSuffix(proxyUrl, "/")

	if !strings.HasPrefix(proxyUrl, "http://") &&
		!strings.HasPrefix(proxyUrl, "https://") {
		proxyUrl = fmt.Sprintf("%s://%s", defaultScheme, proxyUrl)
	}

	u, err := url.Parse(proxyUrl)
	if err != nil {
		return ""
	}
	scheme := fmt.Sprintf("%s://", u.Scheme)
	if strings.HasPrefix(proxyUrl, scheme) {
		s := strings.Replace(proxyUrl, scheme, "", 1)
		s = url.PathEscape(s)
		return scheme + s
	}
	return ""
}
