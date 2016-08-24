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

package main

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

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
)

func main() {
	token := os.Getenv("GH_TOKEN")
	var tc *http.Client
	if len(token) > 0 {
		fmt.Println("Using GH_TOKEN environment variable")
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc = oauth2.NewClient(oauth2.NoContext, ts)
	}
	client := github.NewClient(tc)
	var (
		release *github.RepositoryRelease
		resp    *github.Response
		err     error
	)
	if len(os.Args) > 1 {
		release, resp, err = client.Repositories.GetReleaseByTag("openshift", "origin", os.Args[1])
	} else {
		release, resp, err = client.Repositories.GetLatestRelease("openshift", "origin")
	}
	if err != nil {
		fmt.Printf("Could not get latest OpenShift release: %s", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	assetID := getOpenShiftServerAssetID(release)
	if assetID == 0 {
		fmt.Println("Could not get OpenShift release URL")
		os.Exit(1)
	}
	asset, url, err := client.Repositories.DownloadReleaseAsset("openshift", "origin", assetID)
	if err != nil {
		fmt.Printf("Could not download OpenShift release asset: %s\n", err)
		os.Exit(1)
	}
	if len(url) > 0 {
		httpResp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Could not download OpenShift release asset: %s\n", err)
			os.Exit(1)
		}
		asset = httpResp.Body
	}

	defer asset.Close()

	gzf, err := gzip.NewReader(asset)
	if err != nil {
		fmt.Printf("Could not ungzip OpenShift release asset: %s\n", err)
		os.Exit(1)
	}
	defer gzf.Close()
	tr := tar.NewReader(gzf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			fmt.Printf("Could not extract OpenShift release asset: %s\n", err)
			os.Exit(1)
		}
		if hdr.Typeflag != tar.TypeReg || filepath.Base(hdr.Name) != "kube-apiserver" {
			continue
		}
		contents, err := ioutil.ReadAll(tr)
		if err != nil {
			fmt.Printf("Could not extract OpenShift release asset: %s\n", err)
			os.Exit(1)
		}
		err = ioutil.WriteFile("out/openshift", contents, os.ModePerm)
		if err != nil {
			fmt.Printf("Could not write OpenShift binary: %s\n", err)
			os.Exit(1)
		}
	}
}

func getOpenShiftServerAssetID(release *github.RepositoryRelease) int {
	for _, asset := range release.Assets {
		if strings.HasPrefix(*asset.Name, "openshift-origin-server") && strings.HasSuffix(*asset.Name, "linux-64bit.tar.gz") {
			return *asset.ID
		}
	}
	return 0
}
