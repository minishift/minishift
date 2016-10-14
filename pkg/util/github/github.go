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
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
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

	assetID, filename := getOpenShiftServerAssetID(release)
	if assetID == 0 {
		return errors.New("Could not get OpenShift release URL")
	}
	var asset io.Reader
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

	hasher := sha256.New()

	asset = io.TeeReader(asset, hasher)

	gzf, err := gzip.NewReader(asset)
	if err != nil {
		return errors.Wrap(err, "Could not ungzip OpenShift release asset")
	}
	defer func() { _ = gzf.Close() }()
	destDir := filepath.Dir(outputPath)
	err = os.MkdirAll(destDir, 0755)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "Could not create target directory")
	}
	destFileName := filepath.Base(outputPath)
	tmp, err := ioutil.TempFile(destDir, ".tmp-"+destFileName)
	tr := tar.NewReader(gzf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			os.Remove(tmp.Name())
			return errors.Wrap(err, "Could not extract OpenShift release asset")
		}
		if hdr.Typeflag != tar.TypeReg || filepath.Base(hdr.Name) != "kube-apiserver" {
			continue
		}
		_, err = io.Copy(tmp, tr)
		if err != nil {
			os.Remove(tmp.Name())
			return errors.Wrap(err, "Could not extract OpenShift release asset")
		}
	}
	hash := hex.EncodeToString(hasher.Sum(nil))
	downloadedHash, err := downloadHash(release, filename)
	if err != nil {
		os.Remove(tmp.Name())
		return errors.Wrap(err, "Failed to download hash")
	}
	if len(downloadedHash) == 0 {
		os.Remove(tmp.Name())
		return errors.New("File has no hash to validate - not downloading")
	}

	if hash != downloadedHash {
		os.Remove(tmp.Name())
		return errors.Errorf("Failed to validate hash - expected: %s, actual: %s", hash, downloadedHash)
	}

	err = os.Rename(tmp.Name(), outputPath)
	if err != nil {
		os.Remove(tmp.Name())
		return errors.Wrap(err, "Could not write OpenShift binary")
	}
	return nil
}

func getOpenShiftServerAssetID(release *github.RepositoryRelease) (int, string) {
	for _, asset := range release.Assets {
		if strings.HasPrefix(*asset.Name, "openshift-origin-server") && strings.HasSuffix(*asset.Name, "linux-64bit.tar.gz") {
			return *asset.ID, *asset.Name
		}
	}
	return 0, ""
}

func downloadHash(release *github.RepositoryRelease, filename string) (string, error) {
	assetID := getOpenShiftChecksumAssetID(release)
	if assetID == 0 {
		return "", errors.New("Could not get OpenShift release checksum URL")
	}
	var asset io.Reader
	asset, url, err := client.Repositories.DownloadReleaseAsset("openshift", "origin", assetID)
	if err != nil {
		return "", errors.Wrap(err, "Could not download OpenShift release checksum asset")
	}
	if len(url) > 0 {
		fmt.Printf("Downloading OpenShift %s checksums\n", *release.TagName)
		httpResp, err := http.Get(url)
		if err != nil {
			return "", errors.Wrap(err, "Could not download OpenShift release checksum asset")
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

	scanner := bufio.NewScanner(asset)
	for scanner.Scan() {
		spl := strings.Fields(scanner.Text())
		if len(spl) == 2 && spl[1] == filename {
			return spl[0], nil
		}
	}
	return "", nil
}

func getOpenShiftChecksumAssetID(release *github.RepositoryRelease) int {
	for _, asset := range release.Assets {
		if *asset.Name == "CHECKSUM" {
			return *asset.ID
		}
	}
	return 0
}
