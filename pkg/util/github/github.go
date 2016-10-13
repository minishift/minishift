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
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"gopkg.in/cheggaaa/pb.v1"
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

func DownloadOpenShiftRelease(version, outputPath string) error {
	client := Client()
	var (
		err     error
		release *github.RepositoryRelease
		resp    *github.Response
	)
	if len(version) > 1 {
		release, resp, err = client.Repositories.GetReleaseByTag("openshift", "origin", version)
	} else {
		release, resp, err = client.Repositories.GetLatestRelease("openshift", "origin")
	}
	if err != nil {
		return errors.Wrapf(err, "Could not get OpenShift release")
	}
	defer func() { _ = resp.Body.Close() }()

	assetID := getOpenShiftServerAssetID(release)
	if assetID == 0 {
		return errors.New("Could not get OpenShift release URL")
	}
	asset, url, err := client.Repositories.DownloadReleaseAsset("openshift", "origin", assetID)
	if err != nil {
		return errors.Wrap(err, "Could not download OpenShift release asset")
	}
	if len(url) > 0 {
		fmt.Printf("Downloading OpenShift %s\n", *release.TagName)
		httpResp, err := http.Get(url)
		if err != nil {
			return errors.Wrap(err, "Could not download OpenShift release asset")
		}
		defer func() { _ = httpResp.Body.Close() }()

		asset = httpResp.Body
		if httpResp.ContentLength > 0 {
			bar := pb.New64(httpResp.ContentLength).SetUnits(pb.U_BYTES)
			bar.Start()
			asset = bar.NewProxyReader(asset)
			defer func() {
				<-time.After(bar.RefreshRate)
				fmt.Println()
			}()
		}
	}

	gzf, err := gzip.NewReader(asset)
	if err != nil {
		return errors.Wrap(err, "Could not ungzip OpenShift release asset")
	}
	defer func() { _ = gzf.Close() }()
	tr := tar.NewReader(gzf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return errors.Wrap(err, "Could not extract OpenShift release asset")
		}
		if hdr.Typeflag != tar.TypeReg || filepath.Base(hdr.Name) != "kube-apiserver" {
			continue
		}
		contents, err := ioutil.ReadAll(tr)
		if err != nil {
			return errors.Wrap(err, "Could not extract OpenShift release asset")
		}
		err = os.MkdirAll(filepath.Dir(outputPath), 0755)
		if err != nil && !os.IsExist(err) {
			return errors.Wrap(err, "Could not create target directory")
		}
		err = ioutil.WriteFile(outputPath, contents, os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "Could not write OpenShift binary")
		}
	}
	return nil
}

func getOpenShiftServerAssetID(release *github.RepositoryRelease) int {
	for _, asset := range release.Assets {
		if strings.HasPrefix(*asset.Name, "openshift-origin-server") && strings.HasSuffix(*asset.Name, "linux-64bit.tar.gz") {
			return *asset.ID
		}
	}
	return 0
}
