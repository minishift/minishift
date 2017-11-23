/*
Copyright (C) 2016 Red Hat, Inc.

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

package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type validationTest struct {
	value     string
	shouldErr bool
}

func runValidations(t *testing.T, tests []validationTest, name string, f func(string, string) error) {
	for _, tt := range tests {
		err := f(name, tt.value)
		if !tt.shouldErr {
			assert.NoError(t, err, fmt.Sprintf("Error for testcase %v", tt))
		}
		if tt.shouldErr {
			assert.Error(t, err, fmt.Sprintf("Error for testcase %v", tt))
		}
	}
}

func TestDriver(t *testing.T) {

	var tests = []validationTest{
		{
			value:     "vkasdhfasjdf",
			shouldErr: true,
		},
		{
			value:     "",
			shouldErr: true,
		},
	}

	runValidations(t, tests, "vm-driver", IsValidDriver)
}

func TestValidCIDR(t *testing.T) {
	var tests = []validationTest{
		{
			value:     "0.0.0.0/0",
			shouldErr: false,
		},
		{
			value:     "1.1.1.1/32",
			shouldErr: false,
		},
		{
			value:     "192.168.0.0/16",
			shouldErr: false,
		},
		{
			value:     "255.255.255.255/1",
			shouldErr: false,
		},
		{
			value:     "8.8.8.8/33",
			shouldErr: true,
		},
		{
			value:     "12.1",
			shouldErr: true,
		},
		{
			value:     "1",
			shouldErr: true,
		},
		{
			value:     "a string!",
			shouldErr: true,
		},
		{
			value:     "192.168.1.1/8/",
			shouldErr: true,
		},
	}

	runValidations(t, tests, "cidr", IsValidCIDR)
}

func TestValidURL(t *testing.T) {

	var tests = []validationTest{
		{
			value:     "",
			shouldErr: true,
		},
		{
			value:     "http/foo.com/minishift.tar.gz",
			shouldErr: true,
		},
		{
			value:     "foo/download/minishift.tar.gz",
			shouldErr: true,
		},
		{
			value:     "/absolute/path/no/protocol/minishift.tar.gz",
			shouldErr: false,
		},
		{
			value:     "http://foo.com/minishift.tar.gz",
			shouldErr: false,
		},
		{
			value:     "file:///foo/download/minishift.tar.gz",
			shouldErr: false,
		},
		{
			value:     "centos",
			shouldErr: false,
		},
		{
			value:     "b2d",
			shouldErr: false,
		},
		{
			value:     "random",
			shouldErr: true,
		},
	}
	runValidations(t, tests, "iso-url", IsValidUrl)
}

func TestValidProxyURL(t *testing.T) {

	var tests = []validationTest{
		{
			value:     "",
			shouldErr: false,
		},
		{
			value:     "http://foo.com:3128",
			shouldErr: false,
		},
		{
			value:     "http://127.0.0.1:3128",
			shouldErr: false,
		},
		{
			value:     "http://foo:bar@test.com:324",
			shouldErr: false,
		},
		{
			value:     "https://foo:bar@test.com:454",
			shouldErr: false,
		},
		{
			value:     "https://foo:b@r@test.com:454",
			shouldErr: false,
		},
		{
			value:     "http://myuser:my%20pass@foo.com:3128",
			shouldErr: false,
		},
		{
			value:     "htt://foo.com:3128",
			shouldErr: true,
		},
		{
			value:     "http://:foo.com:3128",
			shouldErr: true,
		},
		{
			value:     "http://myuser@my pass:foo.com:3128",
			shouldErr: true,
		},
		{
			value:     "http://foo:bar@test.com:abc",
			shouldErr: true,
		},
	}
	runValidations(t, tests, "iso-url", IsValidProxy)
}

func TestValidIPv4Address(t *testing.T) {

	var tests = []validationTest{
		{
			value:     "",
			shouldErr: true,
		},
		{
			value:     "10.0.75.128",
			shouldErr: false,
		},
		{
			value:     "localhost",
			shouldErr: true,
		},
		{
			value:     "::1",
			shouldErr: true,
		},
		{
			value:     "fe80::200:5aee:feaa:20a2",
			shouldErr: true,
		},
	}
	runValidations(t, tests, "ipaddress", IsValidIPv4Address)
}

func TestValidNetmask(t *testing.T) {

	var tests = []validationTest{
		{
			value:     "",
			shouldErr: true,
		},
		{
			value:     "0",
			shouldErr: true,
		},
		{
			value:     "1",
			shouldErr: false,
		},
		{
			value:     "24",
			shouldErr: false,
		},
		{
			value:     "25",
			shouldErr: false,
		},
		{
			value:     "26",
			shouldErr: true,
		},
		{
			value:     "255.255.255.0",
			shouldErr: false,
		},
		{
			value:     "255.255.255.128",
			shouldErr: false,
		},
		{
			value:     "128.0.0.0",
			shouldErr: false,
		},
		{
			value:     "255.255.225.0",
			shouldErr: true,
		},
		{
			value:     "127.0.0.0",
			shouldErr: true,
		},
	}
	runValidations(t, tests, "netmask", IsValidNetmask)
}
