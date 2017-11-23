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

package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_interpolate_cmd(t *testing.T) {
	context := NewInterpolationContext()

	context.AddToContext("routing-suffix", "10.0.0.42.xip.io")
	cmd := "ssh sudo mkdir -p /etc/docker/certs.d/docker-registry-default.#{routing-suffix} && sudo ls /etc/docker/certs.d"

	expectedCmd := "ssh sudo mkdir -p /etc/docker/certs.d/docker-registry-default.10.0.0.42.xip.io && sudo ls /etc/docker/certs.d"
	actualCmd := context.Interpolate(cmd)

	assert.Equal(t, expectedCmd, actualCmd)
}

func Test_adding_multiple_interpolation_variables(t *testing.T) {
	context := NewInterpolationContext()

	context.AddToContext("foo", "bar")
	context.AddToContext("snafu", "foobar")
	cmd := "#{foo} #{snafu}"

	expectedCmd := "bar foobar"
	actualCmd := context.Interpolate(cmd)

	assert.Equal(t, expectedCmd, actualCmd)
}

func Test_dollar_sign_in_replacement_value_prints_literally(t *testing.T) {
	context := NewInterpolationContext()

	context.AddToContext("foo", "$bar")
	cmd := "#{foo}"

	expectedCmd := "$bar"
	actualCmd := context.Interpolate(cmd)

	assert.Equal(t, expectedCmd, actualCmd)
}

func Test_adding_and_removing_context_variables(t *testing.T) {
	context := NewInterpolationContext()

	context.AddToContext("foo", "bar")

	expected := "bar"
	actual := context.Interpolate("#{foo}")

	assert.Equal(t, expected, actual)

	context.RemoveFromContext("foo")

	expected = "#{foo}"
	actual = context.Interpolate("#{foo}")
	assert.Equal(t, expected, actual)
}
