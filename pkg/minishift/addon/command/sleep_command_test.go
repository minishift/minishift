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
	"time"
)

func Test_parsing_sleep_time_is_successful(t *testing.T) {
	sleep := NewSleepCommand("sleep 100")

	duration, err := sleep.getSleepTime()

	if err != nil {
		t.Fatal(fmt.Sprintf("Unexpected error '%s'", err.Error()))
	}

	if duration != time.Duration(100)*time.Second {
		t.Fatal(fmt.Sprintf("Unexpected duration: %s", duration))
	}
}

func Test_parsing_sleep_time_ignores_text_after_durationb(t *testing.T) {
	cmd := "sleep 5 foobar"
	sleep := NewSleepCommand(cmd)

	duration, err := sleep.getSleepTime()

	if err != nil {
		t.Fatal(fmt.Sprintf("Unexpected error '%s'", err.Error()))
	}

	if duration != time.Duration(5)*time.Second {
		t.Fatal(fmt.Sprintf("Unexpected duration: %s", duration))
	}
}

func Test_parsing_sleep_time_fails(t *testing.T) {
	cmd := "sleep abc"
	sleep := NewSleepCommand(cmd)

	_, err := sleep.getSleepTime()
	if err == nil {
		t.Fatal("An error should have been returned")
	}

	expectedError := fmt.Sprintf(invalidSleepTimeError, cmd)
	if err.Error() != expectedError {
		t.Fatal(fmt.Sprintf("Unexpected error message. Got '%s'. Expected '%s'", err.Error(), expectedError))
	}
}
