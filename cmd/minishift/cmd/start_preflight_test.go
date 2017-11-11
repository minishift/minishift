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

package cmd

type testData struct {
	in  string
	out bool
}

var sharedIsoURLChecks = []testData{
	// valid URLs (based on protocol only)
	{"http://github.com/minishift/minishift-centos/minishift-centos.iso", true},
	{"https://github.com/minishift/minishift/minishift-b2d.iso", true},

	// invalid URLs
	{"http:/github.com/minishift/minishift.iso", false},

	// wrong file protocol and non existent paths
	{"file:/home/gopher/minishift-b2d.iso", false},
	{"file://home/bar/baz.iso", false},
	{"file://minishift/hoxier/foo", false},

	// no protocol
	{"ram/rahim/anthony.iso", false},
	{"blahblah/http://b2d.iso", false},
	{"blabityblah", false},
	{"/home/joey/chandler/iso.iso", false},
}
