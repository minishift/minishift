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

package github

import (
	"net/http"
	"os"
	"sync"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	client *github.Client

	once sync.Once
)

func Client() *github.Client {
	once.Do(func() {
		token := getToken()
		var tc *http.Client
		if len(token) > 0 {
			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			)
			tc = oauth2.NewClient(oauth2.NoContext, ts)
		}
		client = github.NewClient(tc)
	})
	return client
}

var tokenEnvVars = []string{"MINISHIFT_GITHUB_API_TOKEN", "HOMEBREW_GITHUB_API_TOKEN", "GH_TOKEN"}

func getToken() string {
	for _, envVar := range tokenEnvVars {
		token := os.Getenv(envVar)
		if len(token) > 0 {
			return token
		}
	}
	return ""
}
