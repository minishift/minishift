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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIsoPath(t *testing.T) {
	testData := []struct {
		provided string
		expected string
	}{
		{"https://storage.googleapis.com/minikube/iso/minikube-v0.24.0.iso", filepath.Join("minikube", "v0.24.0")},
		{"https://storage.googleapis.com/minikube/iso/minikube-v0.22.11.iso", filepath.Join("minikube", "v0.22.11")},
		{"https://github.com/minishift/minishift-b2d-iso/releases/download/v1.1.0/minishift-b2d.iso", filepath.Join("b2d", "v1.1.0")},
		{"https://github.com/minishift/minishift-b2d-iso/releases/download/v1.2.0/minishift-b2d.iso", filepath.Join("b2d", "v1.2.0")},
		{"https://github.com/minishift/minishift-centos-iso/releases/download/v1.1.0/minishift-centos7.iso", filepath.Join("centos", "v1.1.0")},
		{"https://github.com/minishift/minishift-centos-iso/releases/download/v1.3.0/minishift-centos7.iso", filepath.Join("centos", "v1.3.0")},
		{"https://foo/v1.2.0/minishift-foo.iso", "unnamed"},
	}

	for _, v := range testData {
		got := GetIsoPath(v.provided)
		assert.Equal(t, v.expected, got)
	}
}

func Test_getMinikubeIsoVersion(t *testing.T) {
	testData := []struct {
		provided string
		expected string
	}{
		{"minikube-v0.24.0.iso", "v0.24.0"},
		{"minikube-v0.22.11.iso", "v0.22.11"},
		{"minishift-v0.22.iso", "v0.22."},
		{"minishift-b2d.iso", ""},
		{"minikube/iso/minikube-v0.24.0.iso", "v0.24.0"},
		{"minishift-b2d-iso/releases/download/v1.2.0/minishift-b2d.iso", "v1.2.0"},
		{"minishift-centos-iso/releases/download/v1.1.0/minishift-centos7.iso", "v1.1.0"},
	}

	for _, v := range testData {
		got := getIsoVersion(v.provided)
		assert.Equal(t, v.expected, got)
	}
}
