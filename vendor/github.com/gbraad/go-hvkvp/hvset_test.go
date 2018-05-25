/*
Copyright (C) 2017 Gerard Braad <me@gbraad.nl>

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

package hvkvp

import (
	"fmt"
	"strings"
	"testing"
)

func assertContains(t *testing.T, a string, b string) {
	if strings.Contains(a, b) {
		return
	}
	t.Fatal(fmt.Sprintf("'%v' doesn't contain '%v'", a, b))
}

func TestPrepareKeyValuePairCommand(t *testing.T) {
	record := NewMachineKVPRecord("test", "Greeting", "Hello, World!")
	command := prepareKeyValuePairCommand(record)

	assertContains(t, command, fmt.Sprintf("kvpDataItem.Name = '%s'", record.Key))
	assertContains(t, command, fmt.Sprintf("kvpDataItem.Data = '%s'", record.Value))
	assertContains(t, command, fmt.Sprintf("kvpDataItem.Source = '%d'", record.Pool))
}
