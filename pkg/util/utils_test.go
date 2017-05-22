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
	"github.com/pkg/errors"
	"testing"
)

// Returns a function that will return n errors, then return successfully forever.
func errorGenerator(n int) func() error {
	generatedErrors := 0
	return func() (err error) {
		if generatedErrors < n {
			generatedErrors += 1
			return errors.New("Error!")
		}
		return nil
	}
}

func TestErrorGenerator(t *testing.T) {
	errorCount := 3
	f := errorGenerator(errorCount)
	for i := 0; i < errorCount-1; i++ {
		if err := f(); err == nil {
			t.Fatalf("Error should have been reported at iteration %v", i)
		}
	}
	if err := f(); err == nil {
		t.Fatal("Error should not have been reported by this call.")
	}
}

func TestRetry(t *testing.T) {

	f := errorGenerator(4)
	if err := Retry(5, f); err != nil {
		t.Fatal("Error should not have been reported during retry.")
	}

	f = errorGenerator(5)
	if err := Retry(4, f); err == nil {
		t.Fatal("Error should have been reported during retry.")
	}

}

func TestEscapeSingleQuote(t *testing.T) {
	experimentstring := map[string]string{
		`foo'bar`:  `foo'"'"'bar`,
		`foo''bar`: `foo'"'"''"'"'bar`,
		`foobar`:   `foobar`,
	}
	for pass, expected := range experimentstring {
		if EscapeSingleQuote(pass) != expected {
			t.Fatalf("Expected '%s' Got '%s'", expected, EscapeSingleQuote(pass))
		}
	}
}

func TestMultiError(t *testing.T) {
	m := MultiError{}

	m.Collect(errors.New("Error 1"))
	m.Collect(errors.New("Error 2"))

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
		t.Fatal("Expected 'false' Got 'true'")
	}
}
