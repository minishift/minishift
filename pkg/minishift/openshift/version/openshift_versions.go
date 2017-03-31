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

package version

import (
	"fmt"
	"io"

	"github.com/google/go-github/github"
	"github.com/minishift/minishift/pkg/minishift/util"
	githubutil "github.com/minishift/minishift/pkg/util/github"
	"github.com/pkg/errors"
)

const githubOwner = "openshift"
const githubRepo = "origin"

func PrintOpenShiftVersionsFromGitHub(output io.Writer) {
	PrintOpenShiftVersions(output)
}

func PrintOpenShiftVersions(output io.Writer) {
	versions, err := getVersions()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Fprint(output, "The following OpenShift versions are available: \n")

	for _, version := range versions {
		if util.ValidateOpenshiftMinVersion(*version.TagName) {
			fmt.Fprintf(output, "\t- %s\n", *version.TagName)
		}
	}
}

func getVersions() ([]*github.RepositoryRelease, error) {
	client := githubutil.Client()

	releases, resp, err := client.Repositories.ListReleases(githubOwner, githubRepo, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if len(releases) == 0 {
		return nil, errors.New("There are no OpenShift versions available.")
	}
	return releases, nil
}
