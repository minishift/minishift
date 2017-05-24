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
)

func Test_default_no_proxy_list(t *testing.T) {
	proxyConfig, err := NewProxyConfig("http://foobar.com", "", "")
	if err != nil {
		t.Fatal("Unexpected error creating proxy config")
	}

	if proxyConfig.NoProxy() != "localhost,127.0.0.1,"+OpenShiftRegistryIp {
		t.Fatalf("Unexpected default no proxy list: %s", proxyConfig.NoProxy())
	}
}

func Test_proxy_config_as_slice(t *testing.T) {
	proxyConfig, err := NewProxyConfig("http://foobar.com", "https://snafu.de", "")
	if err != nil {
		t.Fatal("Unexpected error creating proxy config")
	}

	expectedConfig := []string{"HTTP_PROXY=http://foobar.com", "HTTPS_PROXY=https://snafu.de", "NO_PROXY=localhost,127.0.0.1,172.30.1.1"}
	actualConfig := proxyConfig.ProxyConfig()

	if len(actualConfig) != len(expectedConfig) {
		t.Fatal("Expected and actual config length differ")
	}

	for i, actualValue := range actualConfig {
		if actualValue != expectedConfig[i] {
			t.Fatalf("Expected '%s', got '%s'", expectedConfig[i], actualValue)
		}
	}
}

func Test_set_to_environment(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	proxyConfig, err := NewProxyConfig("http://foobar.com", "https://snafu.de", "42.42.42.42")
	if err != nil {
		t.Fatal("Unexpected error creating proxy config")
	}

	proxyConfig.ApplyToEnvironment()

	expectedValue := "http://foobar.com"
	if os.Getenv("HTTP_PROXY") != expectedValue {
		t.Fatalf("Unexpected proxy setting. Expected '%s'. Got '%s'", expectedValue, os.Getenv("HTTP_PROXY"))
	}

	expectedValue = "https://snafu.de"
	if os.Getenv("HTTPS_PROXY") != expectedValue {
		t.Fatalf("Unexpected proxy setting. Expected '%s'. Got '%s'", expectedValue, os.Getenv("HTTPS_PROXY"))
	}

	expectedValue = "localhost,127.0.0.1,172.30.1.1,42.42.42.42"
	if os.Getenv("NO_PROXY") != expectedValue {
		t.Fatalf("Unexpected proxy setting. Expected '%s'. Got '%s'", expectedValue, os.Getenv("NO_PROXY"))
	}
}

func Test_invalid_http_proxy(t *testing.T) {
	_, err := NewProxyConfig("foo", "", "")
	if err == nil {
		t.Fatal("Expected error, but got none")
	}

	expectedError := "Proxy URL 'foo' is not valid."
	if err.Error() != expectedError {
		t.Fatalf("Unexpected error message, expected '%s', but got '%s'", expectedError, err.Error())
	}

}

func Test_invalid_https_proxy(t *testing.T) {
	_, err := NewProxyConfig("", "bar", "")
	if err == nil {
		t.Fatal("Expected error, but got none")
	}

	expectedError := "Proxy URL 'bar' is not valid."
	if err.Error() != expectedError {
		t.Fatalf("Unexpected error message, expected '%s', but got '%s'", expectedError, err.Error())
	}

}

func Test_add_no_proxy(t *testing.T) {
	proxyConfig, err := NewProxyConfig("http://foobar.com", "https://snafu.de", "42.42.42.42")
	if err != nil {
		t.Fatal("Unexpected error creating proxy config")
	}

	expectedNoProxy := "localhost,127.0.0.1,172.30.1.1,42.42.42.42"
	if expectedNoProxy != proxyConfig.NoProxy() {
		t.Fatalf("Unexpected no proxy list, expected '%s', but got '%s'", expectedNoProxy, proxyConfig.NoProxy())
	}

	proxyConfig.AddNoProxy("snafu.com")
	expectedNoProxy = "localhost,127.0.0.1,172.30.1.1,42.42.42.42,snafu.com"
	if expectedNoProxy != proxyConfig.NoProxy() {
		t.Fatalf("Unexpected no proxy list, expected '%s', but got '%s'", expectedNoProxy, proxyConfig.NoProxy())
	}
}

func Test_validate_proxy_url(t *testing.T) {
	urlList := map[string]bool{
		"": true,
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
		err := ValidateProxyURL(proxyUrl)

		if valid && err != nil {
			t.Errorf("Proxy URL '%s' should be valid, but got error: %s", proxyUrl, err.Error())
		}

		if !valid && err == nil {
			t.Errorf("Proxy URL '%s' should not be valid, but no error received", proxyUrl)
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
		{"HTTP_PROXY", "", false},
		{"HTTPS_PROXY", "", false},
	}

	for _, row := range table {
		os.Setenv(row.envVar, row.envValue)

		proxyConfig, err := NewProxyConfig("", "", "")
		if err != nil {
			t.Errorf("Unexpected error creating proxy config: %v", err)
			continue
		}

		if proxyConfig.IsEnabled() != row.enabled {
			t.Errorf("Expected proxy config to be %t, but got %t", row.enabled, proxyConfig.IsEnabled())
		}

		os.Clearenv()
	}
}
