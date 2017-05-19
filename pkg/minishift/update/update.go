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

package update

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	update "github.com/inconshreveable/go-update"
	"github.com/minishift/minishift/pkg/util/archive"
	pb "gopkg.in/cheggaaa/pb.v1"

	githubutils "github.com/minishift/minishift/pkg/util/github"
	"github.com/minishift/minishift/pkg/version"
)

const (
	timeLayout  = time.RFC1123
	githubOwner = "minishift"
	githubRepo  = "minishift"
)

// CurrentVersion returns the current version of minishift binary installed on the system
func CurrentVersion() (semver.Version, error) {
	localVersion, err := version.GetSemverVersion()
	if err != nil {
		return semver.Version{}, err
	}
	return localVersion, nil
}

// LatestVersion returns the latest version of minishift binary available from upstream
func LatestVersion() (semver.Version, error) {
	latestVersion, err := getLatestVersionFromGitHub(githubOwner, githubRepo)
	if err != nil {
		return semver.Version{}, err
	}
	return latestVersion, nil
}

// IsNewerVersion compares the local and latest versions and returns a boolean
func IsNewerVersion(localVersion, latestVersion semver.Version) bool {
	if localVersion.Compare(latestVersion) < 0 {
		return true
	} else {
		return false
	}
}

// Update handles the update process by calling functions that download, verify, extract and replace the binary.
// It returns an error if any of these functions fails at any point.
func Update(output io.Writer, latestVersion semver.Version) error {

	var extName string

	if runtime.GOOS == "windows" {
		extName = "zip"
	} else {
		extName = "tgz"
	}

	archiveName := fmt.Sprintf("minishift-%s-%s-%s.%s", latestVersion, runtime.GOOS, runtime.GOARCH, extName)
	downloadLinkFormat := "https://github.com/" + githubOwner + "/" + githubRepo + "/releases/download/v%s/%s"

	fmt.Printf("Updating to version %s\n", latestVersion)

	url := fmt.Sprintf(downloadLinkFormat, latestVersion, archiveName)
	// downloadedAchive is a string that has the path to archive on the disk
	downloadedArchive, err := downloadAndVerifyArchive(url)
	if err != nil {
		return err
	}

	dir := strings.Split(downloadedArchive, "minishift")[0]

	defer os.RemoveAll(dir)

	// Extract and replace the downloaded archive in place of current binary

	binary, err := extractBinary(downloadedArchive)
	if err != nil {
		return err
	}

	err = updateBinary(binary)
	if err != nil {
		return err
	}

	return nil
}

// getLatestVersionFromGitHub gets the latest version of minishift available on GitHub.
// It returns the version and error.
func getLatestVersionFromGitHub(githubOwner, githubRepo string) (semver.Version, error) {
	client := githubutils.Client()
	var (
		release *github.RepositoryRelease
		resp    *github.Response
		err     error
	)
	release, resp, err = client.Repositories.GetLatestRelease(githubOwner, githubRepo)
	if err != nil {
		return semver.Version{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	latestVersionString := release.TagName
	if latestVersionString != nil {
		return semver.Make(strings.TrimPrefix(*latestVersionString, "v"))

	}
	return semver.Version{}, fmt.Errorf("Cannot get release name.")
}

// downloadAndVerifyArchive downloads the archive of latest version of minishift from GitHub, saves it to a temporary location,
// and verifies the checksum of downloaded archive with checksum available on GitHub.
// It returns a string containing path to the downloaded archive.
// It returns an error if any of the steps failed.
func downloadAndVerifyArchive(url string) (string, error) {

	fmt.Printf("Downloading %s\n", url)
	httpResp, err := http.Get(url)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Cannot download archive from: %s", url))
	}
	defer func() { _ = httpResp.Body.Close() }()

	updatedArchive := httpResp.Body
	if httpResp.ContentLength > 0 {
		bar := pb.New64(httpResp.ContentLength).SetUnits(pb.U_BYTES)
		bar.Start()
		updatedArchive = bar.NewProxyReader(updatedArchive)
		defer func() {
			<-time.After(bar.RefreshRate)
			fmt.Println()
		}()
	}

	archiveBytes, err := ioutil.ReadAll(updatedArchive)
	if err != nil {
		return "", errors.New("Unable to read downloaded archive")
	}

	// Create a temporary directory inside minishift directory to store archive contents
	dir, err := ioutil.TempDir("", "download")
	if err != nil {
		return "", errors.New("Could not create a temporary directory to store archive contents")
	}

	var downloadedArchive string
	urlSplit := strings.Split(url, "/")

	downloadedArchive = filepath.Join(dir, urlSplit[len(urlSplit)-1])
	if err := writeArchiveToDisk(downloadedArchive, archiveBytes); err != nil {
		return "", err
	}

	// Download and verify checksum
	u := fmt.Sprintf(url + ".sha256")
	checksumResp, err := http.Get(u)
	if err != nil {
		return "", err
	}

	// this newline is to separate the two progress bars for archive and checksum downloads
	fmt.Println()

	defer func() { _ = checksumResp.Body.Close() }()

	checksum := checksumResp.Body
	if checksumResp.ContentLength > 0 {
		bar := pb.New64(checksumResp.ContentLength).SetUnits(pb.U_BYTES)
		bar.Start()
		checksum = bar.NewProxyReader(checksum)
		defer func() {
			<-time.After(2 * bar.RefreshRate)
			fmt.Println()
		}()
	}
	b, err := ioutil.ReadAll(checksum)
	if err != nil {
		return "", err
	}
	if checksumResp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("received %d", checksumResp.StatusCode))
	}

	downloadedChecksum, err := hex.DecodeString(strings.TrimSpace(string(b)))
	if err != nil {
		return "", err
	}

	archiveChecksum, err := checksumFor(crypto.SHA256, archiveBytes)
	if err != nil {
		return "", err
	}

	if !bytes.Equal(archiveChecksum, downloadedChecksum) {
		// if checksum doesn't match, delete the downloaded archive
		os.RemoveAll(downloadedArchive)
		return "", errors.New(fmt.Sprintf("Updated file has wrong checksum. Expected: %x, got: %x", archiveChecksum, checksum))
	}

	return downloadedArchive, nil
}

// checksumFor evaluates and returns the checksum for the payload passed to it.
// It returns an error if given hash function is not linked into the binary.
// Check "crypto" package for more info on hash function.
func checksumFor(h crypto.Hash, payload []byte) ([]byte, error) {
	if !h.Available() {
		return nil, errors.New("requested hash function not available")
	}
	hash := h.New()
	hash.Write(payload) // guaranteed not to error
	return hash.Sum([]byte{}), nil
}

// extractBinary extracts the downloaded archive and returns path to the extracted binary file.
// It returns an error if extraction fails.
func extractBinary(downloadedArchive string) (string, error) {
	var binary string
	dir := strings.Split(downloadedArchive, "minishift")[0]
	extract := filepath.Join(dir, "minishift-extract")

	if runtime.GOOS == "windows" {
		// Unzip the downloadedArchive
		err := archive.Unzip(downloadedArchive, extract)
		if err != nil {
			return "", err
		}
		binary = filepath.Join(extract, "minishift.exe")
	} else {
		ungzip := filepath.Join(dir, "minishift.gzip")

		// Ungzip the archive
		err := archive.Ungzip(downloadedArchive, ungzip)
		if err != nil {
			return "", err
		}

		// Untar the tarball
		err = archive.Untar(ungzip, extract)
		if err != nil {
			return "", err
		}
		binary = filepath.Join(extract, "minishift")
	}

	return binary, nil
}

// updateBinary takes the path to the extracted binary which is also the latest binary for minishift available from GitHub.
// It replaces the binary which called `minishift update` with newer version.
// It throws an error if update fails.
func updateBinary(binary string) error {

	binaryFile, err := os.Open(binary)
	if err != nil {
		return err
	}

	err = update.Apply(binaryFile, update.Options{
		Hash: crypto.SHA256,
		// Checksum: checksum,
	})
	if err != nil {
		return err
		err := update.RollbackError(err)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeArchiveToDisk(downloadedArchive string, archiveBytes []byte) error {
	err := ioutil.WriteFile(downloadedArchive, archiveBytes, 0644)
	return err
}
