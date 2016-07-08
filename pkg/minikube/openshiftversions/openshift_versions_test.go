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
	"testing"

	"github.com/google/go-github/github"
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

	url, _ := url.Parse(server.URL)
	githubClient = github.NewClient(nil)
	githubClient.BaseURL = url
	githubClient.UploadURL = url

	releases, err := getVersions()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(releases) != 2 { // TODO(aprindle) change to len(handler....)
		//Check values here as well?  Write eq method?
		t.Fatalf("Expected two OpenShift releases, it was instead %s", len(releases))
	}
}

type URLHandlerNone struct{}

func (h *URLHandlerNone) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestGetVersionsNone(t *testing.T) {
	// test that an error is returned if nothing is returned at the url endpoint
	handler := &URLHandlerNone{}
	server := httptest.NewServer(handler)

	url, _ := url.Parse(server.URL)
	githubClient = github.NewClient(nil)
	githubClient.BaseURL = url
	githubClient.UploadURL = url

	_, err := getVersions()
	if err == nil {
		t.Fatalf("No kubernetes versions were returned from URL but no error was thrown")
	}
}

type URLHandlerMalformed struct{}

func (h *URLHandlerMalformed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, "Malformed JSON")
}

func TestGetVersionsMalformed(t *testing.T) {
	// test that an error is returned if malformed JSON is at the url endpoint
	handler := &URLHandlerMalformed{}
	server := httptest.NewServer(handler)

	url, _ := url.Parse(server.URL)
	githubClient = github.NewClient(nil)
	githubClient.BaseURL = url
	githubClient.UploadURL = url

	_, err := getVersions()
	if err == nil {
		t.Fatalf("Malformed version value was returned from URL but no error was thrown")
	}
}

func TestPrintOpenShiftVersions(t *testing.T) {
	// test that no openshift version text is printed if there are no versions being served
	// TODO(aprindle) or should this be an error?!?!
	handlerNone := &URLHandlerNone{}
	server := httptest.NewServer(handlerNone)

	url, _ := url.Parse(server.URL)
	githubClient = github.NewClient(nil)
	githubClient.BaseURL = url
	githubClient.UploadURL = url

	var outputBuffer bytes.Buffer
	PrintOpenShiftVersions(&outputBuffer)
	if len(outputBuffer.String()) != 0 {
		t.Fatalf("Expected PrintOpenShiftVersions to not output text as there are no versioned served at the current URL but output was [%s]", outputBuffer.String())
	}

	// test that update text is printed if the latest version is greater than the current version
	// k8sVersionsFromURL = "100.0.0-dev"
	version0 := "0.0.0"
	version1 := "1.0.0"
	handlerCorrect := &URLHandlerCorrect{
		Releases: []*github.RepositoryRelease{{TagName: &version0}, {TagName: &version1}},
	}
	server = httptest.NewServer(handlerCorrect)

	url, _ = url.Parse(server.URL)
	githubClient = github.NewClient(nil)
	githubClient.BaseURL = url
	githubClient.UploadURL = url

	PrintOpenShiftVersions(&outputBuffer)
	if len(outputBuffer.String()) == 0 {
		t.Fatalf("Expected PrintOpenShiftVersion to output text as %s versions were served from URL but output was [%s]",
			2, outputBuffer.String()) //TODO(aprindle) change the 2
	}
}
