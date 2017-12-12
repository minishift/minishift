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
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/pkg/errors"
)

// Returns a function that will return n errors, then return successfully forever.
func errorGenerator(n int, retryable bool) func() error {
	errorCount := 0
	return func() (err error) {
		if errorCount < n {
			errorCount += 1
			e := errors.New("Error!")
			if retryable {
				return &RetriableError{Err: e}
			} else {
				return e
			}

		}

		return nil
	}
}

func TestErrorGenerator(t *testing.T) {
	errors := 3
	f := errorGenerator(errors, false)
	for i := 0; i < errors-1; i++ {
		if err := f(); err == nil {
			t.Fatalf("Error should have been thrown at iteration %v", i)
		}
	}
	if err := f(); err == nil {
		t.Fatalf("Error should not have been thrown this call!")
	}
}

func TestRetry(t *testing.T) {
	f := errorGenerator(4, true)
	if err := Retry(5, f); err != nil {
		t.Fatalf("Error should not have been raised by retry.")
	}

	f = errorGenerator(5, true)
	if err := Retry(4, f); err == nil {
		t.Fatalf("Error should have been raised by retry.")
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
	var versionTestData = []struct {
		OpenshiftVersion    string
		MinOpenshiftVersion string
		expectedResult      bool
	}{
		{"v3.6.0", "v3.7.0", true},
		{"v3.4.1.10", "v3.4.1.2", false},
		{"v3.6.0-alpha.1", "v3.6.0-alpha.2", true},
		{"v3.7.1", "v3.7.1", true},
		{"v3.5.5.31.24", "v3.4.2.1", false},
		{"v3.6.173.0.21", "v3.5.5.31.24", false},
		{"v3.6.0-rc1", "v3.6.0-alpha.1", false},
		{"v3.6.0-alpha.1", "v3.6.0-beta.0", true},
	}

	for _, versionTest := range versionTestData {
		if (VersionOrdinal(versionTest.MinOpenshiftVersion) >= VersionOrdinal(versionTest.OpenshiftVersion)) != versionTest.expectedResult {
			t.Errorf("Expected: %s >= %s", versionTest.MinOpenshiftVersion, versionTest.OpenshiftVersion)
		}
	}
}

var durationTests = []struct {
	in   time.Duration
	want string
}{
	{10*time.Second + 555*time.Millisecond, "10.6s"},
	{10*time.Second + 555*time.Millisecond, "10.6s"},
	{10*time.Second + 500*time.Millisecond, "10.5s"},
	{10*time.Second + 499*time.Millisecond, "10.5s"},
	{9*time.Second + 401*time.Millisecond, "9.4s"},
	{9*time.Second + 456*time.Millisecond, "9.46s"},
	{9*time.Second + 445*time.Millisecond, "9.45s"},
	{1 * time.Second, "1s"},
	{859*time.Millisecond + 445*time.Microsecond, "859.4ms"},
	{859*time.Millisecond + 460*time.Microsecond, "859.5ms"},
	{859*time.Microsecond + 100*time.Nanosecond, "900µs"},
	{45 * time.Nanosecond, "45ns"},
}

func TestFriendlyDuration(t *testing.T) {
	for _, tt := range durationTests {
		got := FriendlyDuration(tt.in)
		expected, _ := time.ParseDuration(tt.want)
		if got != expected {
			t.Errorf("Expected %v but got %v", got, expected)
		}
	}
}

func Test_command_executes_successfully_with_absolute_path(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	dummyBinaryPath := filepath.Join(currentDir, "..", "..", "test", "testdata")

	switch os := runtime.GOOS; os {
	case "darwin":
		dummyBinaryPath = filepath.Join(dummyBinaryPath, "dummybinary_darwin")
	case "linux":
		dummyBinaryPath = filepath.Join(dummyBinaryPath, "dummybinary_linux")
	case "windows":
		dummyBinaryPath = filepath.Join(dummyBinaryPath, "dummybinary_windows.exe")
	default:
		t.Fatal("Unpexpected OS")
	}

	testData := []struct {
		command string
		exists  bool
	}{
		{filepath.Join(currentDir, "..", "..", "test", "testdata", "blahh"), false},
		{dummyBinaryPath, true},
	}
	for _, v := range testData {
		got := CommandExecutesSuccessfully(v.command)
		if got != v.exists {
			t.Errorf("Expected %v for executing %s, got %v", v.exists, dummyBinaryPath, got)
		}
	}
}

func Test_command_executes_successfully_with_command_lookup(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	dummyBinaryPath := filepath.Join(currentDir, "..", "..", "test", "testdata")
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", dummyBinaryPath+string(os.PathListSeparator)+origPath)
	defer func() {
		os.Setenv("PATH", origPath)
	}()

	var cmd string
	switch os := runtime.GOOS; os {
	case "darwin":
		cmd = "dummybinary_darwin"
	case "linux":
		cmd = "dummybinary_linux"
	case "windows":
		cmd = "dummybinary_windows.exe"
	default:
		t.Fatal("Unpexpected OS")
	}

	success := CommandExecutesSuccessfully(cmd)
	if !success {
		t.Errorf("Expected execution of %s to succeed", cmd)
	}
}

func Test_command_executes_unsuccessfully_with_command_lookup(t *testing.T) {
	var cmd string
	switch os := runtime.GOOS; os {
	case "darwin":
		cmd = "dummybinary_darwin"
	case "linux":
		cmd = "dummybinary_linux"
	case "windows":
		cmd = "dummybinary_windows.exe"
	default:
		t.Fatal("Unpexpected OS")
	}

	success := CommandExecutesSuccessfully(cmd)
	if success {
		t.Errorf("Expected execution of %s to fail", cmd)
	}
}
