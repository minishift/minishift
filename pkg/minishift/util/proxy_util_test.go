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
	"testing"
)

func TestParseProxyUriWithHttpProxy(t *testing.T) {
	proxyUri := "http://fedora:fedora123@xyz.com:3128"
	expected := []string{"xyz.com", "3128", "fedora", "fedora123"}
	server, serverPort, user, password, _ := ParseProxyUri(proxyUri)
	result := []string{server, serverPort, user, password}
	for i := range expected {
		if expected[i] != result[i] {
			t.Errorf("Result: %s, Expected: %s", result, expected)
		}
	}
}

func TestParseProxyUriWithHttpsProxy(t *testing.T) {
	proxyUri := "https://fedora:fedora123@xyz.com:3128"
	expected := []string{"xyz.com", "3128", "fedora", "fedora123"}
	server, serverPort, user, password, _ := ParseProxyUri(proxyUri)
	result := []string{server, serverPort, user, password}
	for i := range expected {
		if expected[i] != result[i] {
			t.Errorf("Result: %s, Expected: %s", result, expected)
		}
	}
}

func TestParseProxyUriWithHttpUnauthenticatedProxy(t *testing.T) {
	proxyUri := "http://xyz.com:3128"
	expected := []string{"xyz.com", "3128"}
	server, serverPort, _, _, _ := ParseProxyUri(proxyUri)
	result := []string{server, serverPort}
	for i := range expected {
		if expected[i] != result[i] {
			t.Errorf("Result: %s, Expected: %s", result, expected)
		}
	}
}

func TestParseProxyUriWithHttpUnauthenticatedProxyWithUser(t *testing.T) {
	proxyUri := "https://fedora@xyz.com:3128"
	expected := []string{"xyz.com", "3128", "fedora"}
	server, serverPort, user, _, _ := ParseProxyUri(proxyUri)
	result := []string{server, serverPort, user}
	for i := range expected {
		if expected[i] != result[i] {
			t.Errorf("Result: %s, Expected: %s", result, expected)
		}
	}
}

func TestParseProxyUriWithSpecialCharacterInPassword(t *testing.T) {
	proxyUri := "http://fedora:fedora@@123@xyz.com:3128/"
	expected := []string{"xyz.com", "3128", "fedora", "fedora@@123"}
	server, serverPort, user, password, _ := ParseProxyUri(proxyUri)
	result := []string{server, serverPort, user, password}
	for i := range expected {
		if expected[i] != result[i] {
			t.Errorf("Result: %s, Expected: %s", result, expected)
		}
	}
}
