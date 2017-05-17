/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"testing"
)

// Returns a function that will return n errors, then return successfully forever.
func errorGenerator(n int) func() error {
	errors := 0
	return func() (err error) {
		if errors < n {
			errors += 1
			return fmt.Errorf("Error!")
		}
		return nil
	}
}

func TestErrorGenerator(t *testing.T) {
	errors := 3
	f := errorGenerator(errors)
	for i := 0; i < errors-1; i++ {
		if err := f(); err == nil {
			t.Fatalf("Error should have been reported at iteration %v", i)
		}
	}
	if err := f(); err == nil {
		t.Fatalf("Error should not have been reported by this call.")
	}
}

func TestRetry(t *testing.T) {

	f := errorGenerator(4)
	if err := Retry(5, f); err != nil {
		t.Fatalf("Error should not have been reported during retry.")
	}

	f = errorGenerator(5)
	if err := Retry(4, f); err == nil {
		t.Fatalf("Error should have been reported during retry.")
	}

}

func TestValidateProxyURI(t *testing.T) {
	urlList := map[string]bool{
		"http://foo.com:3128":          true,
		"htt://foo.com:3128":           false,
		"http://127.0.0.1:3128":        true,
		"http://foo:bar@test.com:324":  true,
		"https://foo:bar@test.com:454": true,
		"https://foo:b@r@test.com:454": true,
	}
	for uri, val := range urlList {
		if ValidateProxyURI(uri) != val {
			t.Fatalf("Expected '%t' Got '%t'", val, ValidateProxyURI(uri))
		}
	}
}

func TestMultiError(t *testing.T) {
	m := MultiError{}

	m.Collect(fmt.Errorf("Error 1"))
	m.Collect(fmt.Errorf("Error 2"))

	err := m.ToError()
	expected := `Error 1
Error 2`
	if err.Error() != expected {
		t.Fatalf("%s != %s", err, expected)
	}

	m = MultiError{}
	if err := m.ToError(); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestVersionOrdinal(t *testing.T) {
	if VersionOrdinal("v3.4.1.10") < VersionOrdinal("v3.4.1.2") {
		t.Fatalf("Expected 'false' Got 'true'")
	}
}
