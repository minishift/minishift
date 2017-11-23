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

package image

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Are_Images_Cached(t *testing.T) {
	currDir, err := os.Getwd()
	assert.NoError(t, err, "Unable to determine working directory")

	cacheConfig := &ImageCacheConfig{
		HostCacheDir:      filepath.Join(currDir, "testdata"),
		CachedImages:      []string{"openshift/origin:v3.6.0", "openshift/origin-pod:v3.6.0", "openshift/origin-docker-registry:v3.6.0", "openshift/origin-haproxy-router:v3.6.0"},
		Out:               nil,
		ImageMissStrategy: Skip,
	}

	handler := OciImageHandler{}
	allCached := handler.AreImagesCached(cacheConfig)
	assert.True(t, allCached, "According to the index all images should be cached")

	cacheConfig = &ImageCacheConfig{
		HostCacheDir:      filepath.Join(currDir, "testdata"),
		CachedImages:      []string{"foo/bar:v3.6.0"},
		Out:               nil,
		ImageMissStrategy: Skip,
	}

	handler = OciImageHandler{}
	allCached = handler.AreImagesCached(cacheConfig)
	assert.False(t, allCached, "According to the index the image should not be cached")
}

func Test_Get_Docker_Settings(t *testing.T) {
	var envTests = []struct {
		envMap         map[string]string
		dockerSettings *dockerClientConfig
		errorMessage   string
	}{
		{nil, nil, "The provided Docker environment settings are missing the DOCKER_HOST key."},
		{map[string]string{}, nil, "The provided Docker environment settings are missing the DOCKER_HOST key."},
		{map[string]string{"DOCKER_HOST": "foo"}, nil, "The provided Docker environment settings are missing the DOCKER_CERT_PATH key."},
		{map[string]string{"DOCKER_HOST": "foo", "DOCKER_CERT_PATH": "foo"}, nil, "The provided Docker environment settings are missing the DOCKER_TLS_VERIFY key."},
		{map[string]string{"DOCKER_HOST": "foo", "DOCKER_CERT_PATH": "bar", "DOCKER_TLS_VERIFY": "1"}, &dockerClientConfig{DockerHost: "foo", DockerCertPath: "bar", DockerTLSVerify: true}, ""},
	}

	for _, envTest := range envTests {
		clientConfig, err := getDockerSettings(envTest.envMap)
		if err != nil {
			if envTest.errorMessage != "" {
				assert.EqualError(t, err, envTest.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		}
		assert.EqualValues(t, envTest.dockerSettings, clientConfig)
	}
}
