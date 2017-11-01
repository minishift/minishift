/*
Copyright (C) 2018 Red Hat, Inc.

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

package hostfolder

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var testOptions = []struct {
	options         string
	expectedOptions map[string]string
}{
	{"", map[string]string{}},
	{"user=foo", map[string]string{"user": "foo"}},
	{"user=foo,password=bar", map[string]string{"user": "foo", "password": "bar"}},
	{"user=foo,password=sna,fu", map[string]string{"user": "foo", "password": "sna,fu"}},
	{"user=foo,password=snafu,", map[string]string{"user": "foo", "password": "snafu,"}},
	{"user=foo,password=sna=fu", map[string]string{"user": "foo", "password": "sna=fu"}},
	{"user=foo,password=sna,fu,domain=WORKGROUP", map[string]string{"user": "foo", "password": "sna,fu", "domain": "WORKGROUP"}},
	{"user=foo,password=sn=a,fu,domain=WORKGROUP", map[string]string{"user": "foo", "password": "sn=a,fu", "domain": "WORKGROUP"}},
}

func Test_get_options(t *testing.T) {
	for _, testOption := range testOptions {
		actualOptions := getOptions(testOption.options)
		assert.Equal(t, actualOptions, testOption.expectedOptions, "The extracted options don't match")
	}
}
