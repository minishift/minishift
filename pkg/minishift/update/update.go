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

// Update handles the update process by downloading, verifying, extracting and
// replacing the binary with latest version from GitHub.
// It returns an error if any of these functions fails at any point.
func Update(latestVersion semver.Version) error {
	var extName string

	// Temporary directory to store archive contents
	tmpDir, err := ioutil.TempDir("", "download")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not create a temporary directory: %s", err))
	}

	if runtime.GOOS == "windows" {
		extName = "zip"
	} else {
		extName = "tgz"
	}

	archiveName := fmt.Sprintf("minishift-%s-%s-%s.%s", latestVersion, runtime.GOOS, runtime.GOARCH, extName)
	downloadLinkFormat := "https://github.com/" + githubOwner + "/" + githubRepo + "/releases/download/v%s/%s"

	url := fmt.Sprintf(downloadLinkFormat, latestVersion, archiveName)
	downloadedArchivePath, err := downloadAndVerifyArchive(url, tmpDir)
	if err != nil {
		return err
	}

	// Extract the downloaded archive
	binaryPath, err := extractBinary(downloadedArchivePath, tmpDir)
	if err != nil {
		return err
	}

	// Replace the existing binary with the binary extracted from the archive
	err = updateBinary(binaryPath)
	if err != nil {
		return err
	}

	return nil
}

// getLatestVersionFromGitHub gets the latest version of minishift available on GitHub.
// It returns the version and error.
func getLatestVersionFromGitHub(githubOwner, githubRepo string) (semver.Version, error) {
	var (
		release *github.RepositoryRelease
		resp    *github.Response
		err     error
	)

	client := githubutils.Client()
	release, resp, err = client.Repositories.GetLatestRelease(githubOwner, githubRepo)
	if err != nil {
		return semver.Version{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	latestVersionString := release.TagName
	if latestVersionString != nil {
		return semver.Make(strings.TrimPrefix(*latestVersionString, "v"))

	}

	return semver.Version{}, errors.New("Cannot get release name.")
}

// downloadAndVerifyArchive downloads the archive of latest minishift version from GitHub
// into a temporary location and verifies the checksum of downloaded archive.
// It returns a string containing path to the downloaded archive.
// It returns an error if any of the steps failed.
func downloadAndVerifyArchive(url, tmpDir string) (string, error) {
	fmt.Printf("Downloading %s\n", url)

	httpResp, err := http.Get(url)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Cannot download archive from %s: %s", url, err))
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
		return "", errors.New(fmt.Sprintf("Unable to read downloaded archive: %s", err))
	}

	// Save to archive file
	urlSplit := strings.Split(url, "/")
	downloadedArchivePath := filepath.Join(tmpDir, urlSplit[len(urlSplit)-1])
	if err := ioutil.WriteFile(downloadedArchivePath, archiveBytes, 0644); err != nil {
		return "", err
	}

	// Find checksum of archive
	archiveChecksum, err := checksumFor(crypto.SHA256, archiveBytes)
	if err != nil {
		return "", err
	}

	// Download checksum file from GitHub
	checksumURL := fmt.Sprintf(url + ".sha256")
	checksumResp, err := http.Get(checksumURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = checksumResp.Body.Close() }()

	checksum := checksumResp.Body
	if checksumResp.ContentLength > 0 {
		// Newline is to separate the two progress bars
		fmt.Println()
		bar := pb.New64(checksumResp.ContentLength).SetUnits(pb.U_BYTES)
		bar.Start()
		checksum = bar.NewProxyReader(checksum)
		defer func() {
			<-time.After(2 * bar.RefreshRate)
		}()
	}

	// Verify checksum
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

	// Compare checksums of downloaded checksum and archive file
	if !bytes.Equal(archiveChecksum, downloadedChecksum) {
		// Delete the downloaded archive if checksum doesn't match
		os.RemoveAll(downloadedArchivePath)
		return "", errors.New(fmt.Sprintf("Updated file has wrong checksum. Expected: %x, got: %x", archiveChecksum, checksum))
	}

	return downloadedArchivePath, nil
}

// checksumFor evaluates and returns the checksum for the payload passed to it.
// It returns an error if given hash function is not linked into the binary.
// Check "crypto" package for more info on hash function.
func checksumFor(h crypto.Hash, payload []byte) ([]byte, error) {
	if !h.Available() {
		return nil, errors.New("Requested hash function not available")
	}

	hash := h.New()
	hash.Write(payload) // Guaranteed not to error

	return hash.Sum([]byte{}), nil
}

// extractBinary extracts the downloaded archive and returns path to the extracted binary file.
// It returns an error if extraction fails.
func extractBinary(downloadedArchivePath, archiveDir string) (string, error) {
	var binaryPath string

	extract := filepath.Join(archiveDir, "minishift-extract")
	if runtime.GOOS == "windows" {
		err := archive.Unzip(downloadedArchivePath, extract)
		if err != nil {
			return "", err
		}
		binaryPath = filepath.Join(extract, "minishift.exe")
	} else {
		ungzip := filepath.Join(archiveDir, "minishift.gzip")
		err := archive.Ungzip(downloadedArchivePath, ungzip)
		if err != nil {
			return "", err
		}

		err = archive.Untar(ungzip, extract)
		if err != nil {
			return "", err
		}

		binaryPath = filepath.Join(extract, "minishift")
	}

	return binaryPath, nil
}

// updateBinary takes the path to the extracted binary (latest binary) and replaces the binary.
// It returns an error if update fails.
func updateBinary(binary string) error {
	binaryFile, err := os.Open(binary)
	if err != nil {
		return err
	}

	err = update.Apply(binaryFile, update.Options{
		Hash: crypto.SHA256,
	})
	if err != nil {
		rollbackErr := update.RollbackError(err)
		if rollbackErr != nil {
			return errors.New(fmt.Sprintf("Failed to rollback update: %s", rollbackErr))
		}
		return err
	}

	return nil
}
