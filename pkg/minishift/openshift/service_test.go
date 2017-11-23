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

package openshift

import (
	"fmt"
	"sort"
	"testing"

	"github.com/minishift/minishift/pkg/minikube/constants"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	test "github.com/minishift/minishift/pkg/testing"
	"github.com/minishift/minishift/pkg/testing/testdata"
	"github.com/stretchr/testify/assert"
)

//
type ByName []ServiceSpec

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func Test_getServiceSpecs_with_multiple_nodeports_and_routes(t *testing.T) {
	namespace := "bar"

	setup(t, namespace)
	defer teardown()

	got, err := GetServiceSpecs(namespace)

	assert.NoError(t, err, "Error getting service specs")

	expected := []ServiceSpec{{Namespace: namespace, Name: "guestbook-v1-np", URL: []string(nil), NodePort: "32740", Weight: []string(nil)},
		{Namespace: namespace, Name: "guestbook-v2-np", URL: []string(nil), NodePort: "30485", Weight: []string(nil)},
		{Namespace: namespace, Name: "guestbook-v1", URL: []string{"http://guestbook-v1-3002-myproject.192.168.64.2.nip.io", "http://guestbook-myproject.192.168.64.2.nip.io", "http://guestbook-v1-myproject.192.168.64.2.nip.io"}, NodePort: "", Weight: []string{"", "50%", ""}},
		{Namespace: namespace, Name: "guestbook-v2", URL: []string{"http://guestbook-v2-myproject.192.168.64.2.nip.io", "http://guestbook-v2-3002-myproject.192.168.64.2.nip.io", "http://guestbook-myproject.192.168.64.2.nip.io"}, NodePort: "", Weight: []string{"", "", "50%"}}}

	sort.Sort(ByName(got))
	sort.Sort(ByName(expected))

	comapareServiceSpec(t, got, expected)
}

func comapareServiceSpec(t *testing.T, got, expected []ServiceSpec) {
	for i, service := range got {
		sort.Strings(service.URL)
		sort.Strings(expected[i].URL)
		sort.Strings(service.Weight)
		sort.Strings(expected[i].Weight)
		assert.Equal(t, expected[i].Namespace, service.Namespace)

		for j := range service.URL {
			assert.Equal(t, expected[i].URL[j], service.URL[j])
		}

		assert.Equal(t, expected[i].Name, service.Name)
		assert.Equal(t, expected[i].NodePort, service.NodePort)
		for j := range service.Weight {
			assert.Equal(t, expected[i].Weight[j], service.Weight[j])
		}
	}

}

func setup(t *testing.T, namespace string) {
	instanceState.InstanceConfig = &instanceState.InstanceConfigType{OcPath: "oc"}

	fakeRunner := test.NewFakeRunner(t)
	runner = fakeRunner.Runner

	args := fmt.Sprintf("get projects --config=%s %s", constants.KubeConfigPath, ProjectsCustomCol)
	expectString := fmt.Sprintf("NAME\n%s", namespace)
	fakeRunner.ExpectAndReturn(args, expectString)

	args = fmt.Sprintf("get svc -o json --config=%s -n %s", constants.KubeConfigPath, namespace)
	expectString = testdata.ServiceSpec
	fakeRunner.ExpectAndReturn(args, expectString)

	args = fmt.Sprintf("get route -o json --config=%s -n %s", constants.KubeConfigPath, namespace)
	expectString = testdata.RouteSpec
	fakeRunner.ExpectAndReturn(args, expectString)
}

func teardown() {
	instanceState.InstanceConfig = nil
}
