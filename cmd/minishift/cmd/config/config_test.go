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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type configTestCase struct {
	data   string
	config map[string]interface{}
}

var configTestCases = []configTestCase{
	{
		data: `{
    "memory": 2
}`,
		config: map[string]interface{}{
			"memory": 2,
		},
	},
	{
		data: `{
    "ReminderWaitPeriodInHours": 99,
    "cpus": 4,
    "disk-size": "20g",
    "iso-url": "http://foo.bar/minishift-centos.iso",
    "log_dir": "/etc/hosts",
    "show-libmachine-logs": true,
    "v": 5,
    "vm-driver": "kvm"
}`,
		config: map[string]interface{}{
			"iso-url":   "http://foo.bar/minishift-centos.iso",
			"vm-driver": "kvm",
			"cpus":      4,
			"disk-size": "20g",
			"v":         5,
			"show-libmachine-logs":      true,
			"log_dir":                   "/etc/hosts",
			"ReminderWaitPeriodInHours": 99,
		},
	},
	{
		data: `{
    "host-config-dir": "/etc/foo/bar",
    "host-data-dir": "/etc/foo/bar",
    "host-only-cidr": "/etc/foo/bar",
    "host-pv-dir": "/etc/foo/bar",
    "host-volumes-dir": "/etc/foo/bar"
}`,
		config: map[string]interface{}{
			"host-config-dir":  "/etc/foo/bar",
			"host-data-dir":    "/etc/foo/bar",
			"host-only-cidr":   "/etc/foo/bar",
			"host-pv-dir":      "/etc/foo/bar",
			"host-volumes-dir": "/etc/foo/bar",
		},
	},
	{
		data: `{
    "http-proxy": "http://foo.com:3128",
    "https-proxy": "https://foo.com:3128",
    "logging": false,
    "public-hostname": "localhost",
    "routing-suffix": ".xip.io"
}`,
		config: map[string]interface{}{
			"http-proxy":      "http://foo.com:3128",
			"https-proxy":     "https://foo.com:3128",
			"logging":         false,
			"public-hostname": "localhost",
			"routing-suffix":  ".xip.io",
		},
	},
}

func TestReadConfig(t *testing.T) {
	for _, tt := range configTestCases {
		r := bytes.NewBufferString(tt.data)
		config, err := decode(r)
		assert.NoError(t, err, "Error Decoding config")
		assert.ObjectsAreEqualValues(tt.data, config)
	}
}

func TestWriteConfig(t *testing.T) {
	var b bytes.Buffer
	for _, tt := range configTestCases {
		err := encode(&b, tt.config)
		assert.NoError(t, err, "Error Encoding config")
		assert.Equal(t, b.String(), tt.data)
		b.Reset()
	}
}
