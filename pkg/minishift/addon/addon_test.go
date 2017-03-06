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

package addon

import (
	"fmt"
	"sort"
	"testing"
)

type DummyAddOn struct {
	AddOn

	meta     *DummyAddOnMeta
	priority int
	enabled  bool
}

func (a *DummyAddOn) MetaData() AddOnMeta {
	return a.meta
}

func (a *DummyAddOn) GetPriority() int {
	return a.priority
}

func (a *DummyAddOn) IsEnabled() bool {
	return a.enabled
}

func (a *DummyAddOn) String() string {
	return a.MetaData().Name()
}

type DummyAddOnMeta struct {
	AddOnMeta

	name string
}

func (m *DummyAddOnMeta) Name() string {
	return m.name
}

var a = &DummyAddOn{meta: &DummyAddOnMeta{name: "a"}, enabled: false, priority: 8}
var b = &DummyAddOn{meta: &DummyAddOnMeta{name: "b"}, enabled: true, priority: 5}
var c = &DummyAddOn{meta: &DummyAddOnMeta{name: "c"}, enabled: false, priority: -1}
var d = &DummyAddOn{meta: &DummyAddOnMeta{name: "d"}, enabled: true, priority: 0}
var e = &DummyAddOn{meta: &DummyAddOnMeta{name: "e"}, enabled: true, priority: 0}

func Test_sorting_addons_by_priority(t *testing.T) {
	addOns := []AddOn{a, b, c, d, e}
	sort.Sort(ByPriority(addOns))

	expectedSortOrder := "[c d e b a]"
	if fmt.Sprint(addOns) != expectedSortOrder {
		t.Fatal(fmt.Sprintf("Expected sort order '%s', but got '%s'", expectedSortOrder, fmt.Sprint(addOns)))
	}
}

func Test_sorting_addons_by_state_and_name(t *testing.T) {
	addOns := []AddOn{a, b, c, d, e}
	sort.Sort(ByStatusThenName(addOns))

	expectedSortOrder := "[b d e a c]"
	if fmt.Sprint(addOns) != expectedSortOrder {
		t.Fatal(fmt.Sprintf("Expected sort order '%s', but got '%s'", expectedSortOrder, fmt.Sprint(addOns)))
	}
}

func Test_sorting_addons_by_state_priority_and_name(t *testing.T) {
	addOns := []AddOn{a, b, c, d, e}
	sort.Sort(ByStatusThenPriorityThenName(addOns))

	expectedSortOrder := "[d e b c a]"
	if fmt.Sprint(addOns) != expectedSortOrder {
		t.Fatal(fmt.Sprintf("Expected sort order '%s', but got '%s'", expectedSortOrder, fmt.Sprint(addOns)))
	}
}
