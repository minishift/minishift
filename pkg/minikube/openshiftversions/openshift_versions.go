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
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

const githubOwner = "openshift"
const githubRepo = "origin"

func PrintOpenShiftVersionsFromGitHub(output io.Writer) {
	PrintOpenShiftVersions(output)
}

func PrintOpenShiftVersions(output io.Writer) {
	versions, err := getVersions()
	if err != nil {
		glog.Errorln(err)
		return
	}
	fmt.Fprint(output, "The following OpenShift versions are available: \n")

	for _, version := range versions {
		fmt.Fprintf(output, "\t- %s\n", *version.TagName)
	}
}

var (
	githubClient *github.Client = nil
)

func getVersions() ([]*github.RepositoryRelease, error) {
	if githubClient == nil {
		token := os.Getenv("GH_TOKEN")
		var tc *http.Client
		if len(token) > 0 {
			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			)
			tc = oauth2.NewClient(oauth2.NoContext, ts)
		}
		githubClient = github.NewClient(tc)
	}
	client := githubClient

	releases, resp, err := client.Repositories.ListReleases(githubOwner, githubRepo, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if len(releases) == 0 {
		return nil, fmt.Errorf("There were no OpenShift Releases available")
	}
	return releases, nil
}
