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
	"fmt"
	"testing"
)

func Test_interpolate_cmd(t *testing.T) {
	context := NewInterpolationContext()

	context.AddToContext("routing-suffix", "10.0.0.42.xip.io")
	cmd := "ssh sudo mkdir -p /etc/docker/certs.d/docker-registry-default.#{routing-suffix} && sudo ls /etc/docker/certs.d"

	expectedCmd := "ssh sudo mkdir -p /etc/docker/certs.d/docker-registry-default.10.0.0.42.xip.io && sudo ls /etc/docker/certs.d"
	actualCmd := context.Interpolate(cmd)

	if expectedCmd != actualCmd {
		t.Fatal(fmt.Sprintf("Expected command: %s. Got: %s.", expectedCmd, actualCmd))
	}
}

func Test_adding_invalid_key_value(t *testing.T) {
	context := NewInterpolationContext()

	key := "*+"
	value := ""
	err := context.AddToContext(key, value)

	if err == nil {
		t.Fatal(fmt.Sprintf("It should not have been possible to add %s/%s", key, value))
	}
}

func Test_multiple_contexts(t *testing.T) {
	context := NewInterpolationContext()

	context.AddToContext("foo", "bar")
	context.AddToContext("snafu", "foobar")
	cmd := "#{foo} #{snafu}"

	expectedCmd := "bar foobar"
	actualCmd := context.Interpolate(cmd)

	if expectedCmd != actualCmd {
		t.Fatal(fmt.Sprintf("Expected command: %s. Got: %s.", expectedCmd, actualCmd))
	}
}

func Test_adding_and_removing_context(t *testing.T) {
	context := NewInterpolationContext()

	context.AddToContext("foo", "bar")

	expected := "bar"
	actual := context.Interpolate("#{foo}")

	if expected != actual {
		t.Fatal(fmt.Sprintf("Expected: %s. Got: %s.", expected, actual))
	}

	context.RemoveFromContext("foo")

	expected = "#{foo}"
	actual = context.Interpolate("#{foo}")
	if expected != actual {
		t.Fatal(fmt.Sprintf("Expected: %s. Got: %s.", expected, actual))
	}
}
