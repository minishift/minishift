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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_default_no_proxy_list(t *testing.T) {
	proxyConfig, err := NewProxyConfig("http://foobar.com", "", "")
	assert.NoError(t, err, "Error in getting new proxy config")

	assert.Equal(t, "localhost,127.0.0.1,"+OpenShiftRegistryIp, proxyConfig.NoProxy())
}

func Test_proxy_config_as_slice(t *testing.T) {
	proxyConfig, err := NewProxyConfig("http://foobar.com", "https://snafu.de", "")
	assert.NoError(t, err, "Error in getting new proxy config")

	expectedConfig := []string{"HTTP_PROXY=http://foobar.com",
		"http_proxy=http://foobar.com",
		"HTTPS_PROXY=https://snafu.de",
		"https_proxy=https://snafu.de",
		"NO_PROXY=localhost,127.0.0.1,172.30.1.1",
		"no_proxy=localhost,127.0.0.1,172.30.1.1"}
	actualConfig := proxyConfig.ProxyConfig()

	assert.Len(t, expectedConfig, len(actualConfig))

	for i, actualValue := range actualConfig {
		assert.Equal(t, expectedConfig[i], actualValue)
	}
}

func Test_set_to_environment(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	proxyConfig, err := NewProxyConfig("http://foobar.com", "https://snafu.de", "42.42.42.42")
	assert.NoError(t, err, "Error in getting new proxy config")

	proxyConfig.ApplyToEnvironment()

	expectedValue := "http://foobar.com"
	assert.Equal(t, expectedValue, os.Getenv("HTTP_PROXY"))

	expectedValue = "https://snafu.de"
	assert.Equal(t, expectedValue, os.Getenv("HTTPS_PROXY"))

	expectedValue = "localhost,127.0.0.1,172.30.1.1,42.42.42.42"
	assert.Equal(t, expectedValue, os.Getenv("NO_PROXY"))
}

func Test_add_no_proxy(t *testing.T) {
	proxyConfig, err := NewProxyConfig("http://foobar.com", "https://snafu.de", "42.42.42.42")
	assert.NoError(t, err, "Error in getting new proxy config")

	expectedNoProxy := "localhost,127.0.0.1,172.30.1.1,42.42.42.42"
	assert.Equal(t, expectedNoProxy, proxyConfig.NoProxy())

	proxyConfig.AddNoProxy("snafu.com")
	expectedNoProxy = "localhost,127.0.0.1,172.30.1.1,42.42.42.42,snafu.com"
	assert.Equal(t, expectedNoProxy, proxyConfig.NoProxy())
}

func Test_validate_proxy_url(t *testing.T) {
	urlList := map[string]bool{
		"":                                     true,
		"foo.com:3128":                         true,
		"http://foo.com:3128":                  true, // special case for us as part of ProxyConfig
		"http://127.0.0.1:3128":                true,
		"http://foo:bar@test.com:324":          true,
		"https://foo:bar@test.com:454":         true,
		"https://foo:b@r@test.com:454":         true,
		"http://myuser:my%20pass@foo.com:3128": true,
		"htt://foo.com:3128":                   false,
		"http://:foo.com:3128":                 false,
		"http://myuser@my pass:foo.com:3128":   false,
		"http://foo:bar@test.com:abc":          false,
	}
	for proxyUrl, valid := range urlList {
		err := ValidateProxyURL(proxyUrl, "http")
		if valid {
			assert.NoError(t, err)
		}
		if !valid {
			assert.Error(t, err)
		}
	}
}

func Test_http_proxy_from_env(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	var table = []struct {
		envVar   string
		envValue string
		enabled  bool
	}{
		{"HTTP_PROXY", "http://user:pass@myproxy.foo:1080", true},
		{"HTTPS_PROXY", "http://user:pass@myproxy.foo:1080", true},
		{"http_proxy", "http://user:pass@myproxy.foo:1080", true},
		{"https_proxy", "http://user:pass@myproxy.foo:1080", true},
		{"HTTP_PROXY", "", false},
		{"HTTPS_PROXY", "", false},
	}

	for _, row := range table {
		os.Setenv(row.envVar, row.envValue)

		proxyConfig, err := NewProxyConfig("", "", "")
		assert.NoError(t, err)

		assert.Equal(t, row.enabled, proxyConfig.IsEnabled())

		os.Clearenv()
	}
}

func Test_http_proxy_from_env_lowercase_precedence(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	os.Setenv("HTTP_PROXY", "http://user:pass@myproxy.foo:1080")
	os.Setenv("http_proxy", "http://user:pass@someotherproxy.foo:1080")
	os.Setenv("HTTPS_PROXY", "https://user:pass@myproxy.foo:1080")
	os.Setenv("https_proxy", "https://user:pass@someotherproxy.foo:1080")

	proxyConfig, err := NewProxyConfig("", "", "")
	assert.NoError(t, err)

	assert.Equal(t, "http://user:pass@someotherproxy.foo:1080", proxyConfig.httpProxy)
	assert.Equal(t, "https://user:pass@someotherproxy.foo:1080", proxyConfig.httpsProxy)
}

func Test_parse_special_character_uri(t *testing.T) {
	var urlList = []struct {
		givenURI    string
		expectedURI string
	}{
		{"", ""},
		{"foo.com:3128", "http://foo.com:3128"},
		{"http://foo.com:3128", "http://foo.com:3128"},
		{"http://user:F@oo!B#ar$@myserver:3128", "http://user:F@oo%21B%23ar$@myserver:3128"},
		{"https://myuser:my#pass@foo.com:3128", ""},
		{"https://newuser:new(pas*)wrd@test.com:3128", "https://newuser:new%28pas%2A%29wrd@test.com:3128"},
	}
	for _, proxyUrl := range urlList {
		got := parseProxySpecialChar(proxyUrl.givenURI, "http")
		assert.Equal(t, proxyUrl.expectedURI, got)
	}
}
