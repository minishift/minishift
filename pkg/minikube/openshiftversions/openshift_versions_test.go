/*
Copyright (C) 2016 Red Hat, Inc.

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

package openshiftversions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-github/github"

	githubutils "github.com/minishift/minishift/pkg/util/github"
)

type URLHandlerCorrect struct {
	Releases []*github.RepositoryRelease
}

func (h *URLHandlerCorrect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(h.Releases)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, string(b))
}

func TestGetVersionsCorrect(t *testing.T) {
	// test that the version is correctly parsed if returned if valid JSON is returned the url endpoint
	version0 := "0.0.0"
	version1 := "1.0.0"
	handler := &URLHandlerCorrect{
		Releases: []*github.RepositoryRelease{{TagName: &version0}, {TagName: &version1}},
	}
	server := httptest.NewServer(handler)

	parsedUrl, _ := url.Parse(server.URL)
	os.Clearenv()
	githubClient := githubutils.Client()
	githubClient.BaseURL = parsedUrl
	githubClient.UploadURL = parsedUrl

	releases, err := getVersions()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(releases) != 2 { // TODO(aprindle) change to len(handler....)
		//Check values here as well?  Write eq method?
		t.Fatalf("Expected two OpenShift releases, received instead %s", len(releases))
	}
}

type URLHandlerNone struct{}

func (h *URLHandlerNone) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestGetVersionsNone(t *testing.T) {
	// test that an error is returned if nothing is returned at the url endpoint
	handler := &URLHandlerNone{}
	server := httptest.NewServer(handler)

	parsedUrl, _ := url.Parse(server.URL)
	os.Clearenv()
	githubClient := githubutils.Client()
	githubClient.BaseURL = parsedUrl
	githubClient.UploadURL = parsedUrl

	_, err := getVersions()
	if err == nil {
		t.Fatal("No OpenShift versions returned from URL but no error reported.")
	}
}

type URLHandlerMalformed struct{}

func (h *URLHandlerMalformed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprint(w, "Malformed JSON")
}

func TestGetVersionsMalformed(t *testing.T) {
	// test that an error is returned if malformed JSON is at the url endpoint
	handler := &URLHandlerMalformed{}
	server := httptest.NewServer(handler)

	parsedUrl, _ := url.Parse(server.URL)
	os.Clearenv()
	githubClient := githubutils.Client()
	githubClient.BaseURL = parsedUrl
	githubClient.UploadURL = parsedUrl

	_, err := getVersions()
	if err == nil {
		t.Fatal("Incorrect version value returned from URL but no error reported.")
	}
}

func TestPrintOpenShiftVersions(t *testing.T) {
	// test that no openshift version text is printed if there are no versions being served
	// TODO(aprindle) or should this be an error?!?!
	handlerNone := &URLHandlerNone{}
	server := httptest.NewServer(handlerNone)

	parsedUrl, _ := url.Parse(server.URL)
	os.Clearenv()
	githubClient := githubutils.Client()
	githubClient.BaseURL = parsedUrl
	githubClient.UploadURL = parsedUrl

	var outputBuffer bytes.Buffer
	PrintOpenShiftVersions(&outputBuffer)
	if len(outputBuffer.String()) != 0 {
		t.Fatalf("Expected no output from PrintOpenShiftVersions because no versions exist at the current URL but the output was [%s]", outputBuffer.String())
	}

	// test that update text is printed if the latest version is greater than the current version
	// k8sVersionsFromURL = "100.0.0-dev"
	version0 := "0.0.0"
	version1 := "1.0.0"
	handlerCorrect := &URLHandlerCorrect{
		Releases: []*github.RepositoryRelease{{TagName: &version0}, {TagName: &version1}},
	}
	server = httptest.NewServer(handlerCorrect)

	parsedUrl, _ = url.Parse(server.URL)
	githubClient.BaseURL = parsedUrl
	githubClient.UploadURL = parsedUrl

	PrintOpenShiftVersions(&outputBuffer)
	if len(outputBuffer.String()) == 0 {
		t.Fatalf("Expected PrintOpenShiftVersion to output %d versions from the current URL but the output was [%s]",
			2, outputBuffer.String()) //TODO(aprindle) change the 2
	}
}
