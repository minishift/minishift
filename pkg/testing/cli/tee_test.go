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

package cli

import (
	"fmt"
	"os"
	"testing"
)

func Test_tee_captures_stdout_and_stderr(t *testing.T) {
	tee, err := NewTee(true)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	fmt.Fprint(os.Stdout, "Hello")
	fmt.Fprint(os.Stderr, "world!")

	tee.Close()

	if tee.StdoutBuffer.String() != "Hello" {
		t.Fatalf("Wrong stdout capture: '%s'", tee.StdoutBuffer.String())
	}

	if tee.StderrBuffer.String() != "world!" {
		t.Fatalf("Wrong stderr capture: '%s'", tee.StderrBuffer.String())
	}
}
