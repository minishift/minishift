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
	"bufio"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"gopkg.in/cheggaaa/pb.v1"

	"crypto/sha256"
	"fmt"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/archive"
	minishiftos "github.com/minishift/minishift/pkg/util/os"
	"io/ioutil"
	"path/filepath"
	"regexp"
)

type OpenShiftBinaryType string

const (
	OC        OpenShiftBinaryType = "oc"
	OPENSHIFT OpenShiftBinaryType = "openshift"
)

func (t OpenShiftBinaryType) String() string {
	return string(t)
}

const (
	TAR = "tar.gz"
	ZIP = "zip"
)

var (
	client *github.Client
	once   sync.Once
)

func Client() *github.Client {
	once.Do(func() {
		token := GetGitHubApiToken()
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

func GetGitHubApiToken() string {
	for _, envVar := range tokenEnvVars {
		token := os.Getenv(envVar)
		if len(token) > 0 {
			return token
		}
	}
	return ""
}

func DownloadOpenShiftReleaseBinary(binaryType OpenShiftBinaryType, osType minishiftos.OS, version, outputPath string, proxyUrl string) error {
	client := Client()
	var (
		err     error
		release *github.RepositoryRelease
		resp    *github.Response
	)
	// Get the GitHub release information - either latest or for the specified version
	errorMessage := ""
	if len(version) > 1 {
		release, resp, err = client.Repositories.GetReleaseByTag("openshift", "origin", version)
		errorMessage = fmt.Sprintf("Cannot get the OpenShift release version %s", version)
	} else {
		release, resp, err = client.Repositories.GetLatestRelease("openshift", "origin")
		errorMessage = "Cannot get the latest OpenShift release."

	}
	if err != nil {
		return errors.Wrap(err, errorMessage)
	}

	defer func() { _ = resp.Body.Close() }()

	// Get asset id and filename based on the method parameters
	assetID, assetFilename := getAssetIdAndFilename(binaryType, osType, release)
	if assetID == 0 {
		return errors.New(fmt.Sprintf("Cannot get binary '%s' in version %s for the target environment %s",
			binaryType.String(), version, strings.Title(osType.String())))
	}

	// Download the asset
	var asset io.Reader
	asset, uri, err := client.Repositories.DownloadReleaseAsset("openshift", "origin", assetID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Cannot download OpenShift release asset %d", assetID))
	}
	if len(uri) > 0 {
		fmt.Println(fmt.Sprintf("Downloading OpenShift binary '%s' version '%s'", binaryType.String(), *release.TagName))
		myClient := util.GetHttpClient(proxyUrl)
		httpResp, err := myClient.Get(uri)
		if err != nil {
			return errors.Wrap(err, "Cannot download OpenShift release asset.")
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

	// Create target directory and file
	tmpDir, err := ioutil.TempDir("", "minishift-asset-download-")
	if err != nil {
		return errors.Wrap(err, "Cannot create temporary download directory.")
	}
	defer os.RemoveAll(tmpDir)

	// Create a tmp directory for the asset
	assetTmpFile := filepath.Join(tmpDir, assetFilename)
	out, err := os.Create(assetTmpFile)
	defer out.Close()
	if err != nil {
		return errors.Wrapf(err, "Cannot create file %s", assetTmpFile)
	}

	// Copy the asset and verify its hash
	_, err = io.Copy(out, asset)
	if err != nil {
		return errors.Wrapf(err, "Unexpected error occured while copying %s to %s", assetTmpFile, tmpDir)
	}
	err = out.Sync()
	if err != nil {
		return errors.Wrapf(err, "Unexpected error occured while copying %s to %s", assetTmpFile, tmpDir)
	}

	// Disabling hash verification due to inconsistent checksums on OpenShift download page - https://github.com/openshift/origin/issues/12025
	//hash := hex.EncodeToString(hasher.Sum(nil))
	//downloadedHash, err := downloadHash(release, assetFilename)
	//if err != nil {
	//	return errors.Wrap(err, "Failed to download hash")
	//}
	//if len(downloadedHash) == 0 {
	//	return errors.New("File has no hash to validate - not downloading")
	//}
	//
	//if hash != downloadedHash {
	//	return errors.Errorf("Failed to validate hash - expected: %s, actual: %s", hash, downloadedHash)
	//}

	// Unpack the asset
	binaryPath := ""
	switch {
	case strings.HasSuffix(assetTmpFile, TAR):
		// unzip
		tarFile := assetTmpFile[:len(assetTmpFile)-3]
		err = archive.Ungzip(assetTmpFile, tarFile)
		if err != nil {
			return errors.Wrapf(err, "Cannot ungzip %s", assetTmpFile)
		}

		// untar
		err = archive.Untar(tarFile, tmpDir)
		if err != nil {
			return errors.Wrapf(err, "Cannot untar %s", tarFile)
		}

		content, err := listDirExcluding(tmpDir, ".*.tar.*")
		if err != nil {
			return errors.Wrapf(err, "Cannot list content of %s", tmpDir)
		}
		if len(content) > 1 {
			return errors.New(fmt.Sprintf("Unexpected number of files in tmp directory: %s", content))
		}

		binaryPath = filepath.Join(tmpDir, content[0])
	case strings.HasSuffix(assetTmpFile, ZIP):
		contentDir := assetTmpFile[:len(assetTmpFile)-4]
		err = archive.Unzip(assetTmpFile, contentDir)
		if err != nil {
			return errors.Wrapf(err, "Cannot unzip %s", assetTmpFile)
		}
		binaryPath = contentDir
	}

	binaryName := binaryType.String()
	if osType == minishiftos.WINDOWS {
		binaryName = binaryName + ".exe"
	}
	binaryPath = filepath.Join(binaryPath, binaryName)

	// Copy the requested asset into its final destination
	err = os.MkdirAll(outputPath, 0755)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "Cannot create the target directory.")
	}

	finalBinaryPath := filepath.Join(outputPath, binaryName)
	copy(binaryPath, finalBinaryPath)
	if err != nil {
		return err
	}

	err = os.Chmod(finalBinaryPath, 0777)
	if err != nil {
		return errors.Wrapf(err, "Cannot make %s executable", finalBinaryPath)
	}

	return nil
}

func listDirExcluding(dir string, excludeRegexp string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, f := range files {
		matched, err := regexp.MatchString(excludeRegexp, f.Name())
		if err != nil {
			return nil, err
		}

		if !matched {
			result = append(result, f.Name())
		}

	}

	return result, nil
}

func getAssetIdAndFilename(binaryType OpenShiftBinaryType, osType minishiftos.OS, release *github.RepositoryRelease) (int, string) {
	prefix := ""
	switch binaryType {
	case OC:
		prefix = "openshift-origin-client-tools"
	case OPENSHIFT:
		prefix = "openshift-origin-server"
	default:
		errors.New("Unexpected binary type")
	}

	suffix := ""
	switch osType {
	case minishiftos.LINUX:
		suffix = "linux-64bit.tar.gz"
	case minishiftos.DARWIN:
		suffix = "mac.zip"
	case minishiftos.WINDOWS:
		suffix = "windows.zip"
	default:
		errors.New("Unexpected OS type")
	}

	for _, asset := range release.Assets {
		if strings.HasPrefix(*asset.Name, prefix) && strings.HasSuffix(*asset.Name, suffix) {
			return *asset.ID, *asset.Name
		}
	}
	return 0, ""
}

func downloadHash(release *github.RepositoryRelease, filename string) (string, error) {
	checksumAssetID := getOpenShiftChecksumAssetID(release)
	if checksumAssetID == 0 {
		return "", errors.New("Cannot get the OpenShift release checksum URL.")
	}
	var asset io.Reader
	asset, url, err := client.Repositories.DownloadReleaseAsset("openshift", "origin", checksumAssetID)
	if err != nil {
		return "", errors.Wrap(err, "Cannot download the OpenShift release checksum asset.")
	}
	if len(url) > 0 {
		glog.V(2).Infof("Downloading OpenShift %s checksums\n", *release.TagName)
		httpResp, err := http.Get(url)
		if err != nil {
			return "", errors.Wrap(err, "Cannot download the OpenShift release checksum asset.")
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

func copy(src, dest string) error {
	glog.V(2).Infof("Copying %s to %s\n", src, dest)
	srcFile, err := os.Open(src)
	defer srcFile.Close()
	if err != nil {
		return errors.Wrapf(err, "Cannot open src file %s", src)
	}

	destFile, err := os.Create(dest)
	defer destFile.Close()
	if err != nil {
		return errors.Wrapf(err, "Cannot create dst file %s", dest)
	}

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return errors.Wrapf(err, "Cannot copy %s to %s", src, dest)
	}

	err = destFile.Sync()
	if err != nil {
		return errors.Wrapf(err, "Cannot copy %s to %s", src, dest)
	}

	return nil
}
