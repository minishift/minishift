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
	"testing"
)

func TestGetKvpRecords(t *testing.T) {
	records := getKvpRecords("./testdata/testpool")
	expected := 2
	actual := len(records)

	if actual != expected {
		t.Errorf("Expected '%d' but got '%d'.", expected, actual)
	}
}

func TestGetKvpRecordsByKey(t *testing.T) {
	records := getKvpRecords("./testdata/testpool")

	testData := map[string]string{
		"IpAddress": "10.0.75.128",
		"42":        "Answer to the Ultimate Question of Life, the Universe, and Everything",
	}

	for _, record := range records {
		key := record.GetKey()
		actual := record.GetValue()
		expected := testData[key]

		if actual != expected {
			t.Errorf("Expected '%s' but got '%s' using '%s' as key.", expected, actual, key)
		}
	}
}
