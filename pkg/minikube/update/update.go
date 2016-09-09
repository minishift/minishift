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
	"crypto"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"golang.org/x/oauth2"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"github.com/google/go-github/github"
	update "github.com/inconshreveable/go-update"
	"github.com/jimmidyson/minishift/pkg/minikube/config"
	"github.com/jimmidyson/minishift/pkg/minikube/constants"
	"github.com/jimmidyson/minishift/pkg/version"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"
)

const (
	timeLayout  = time.RFC1123
	githubOwner = "jimmidyson"
	githubRepo  = "minishift"
)

var (
	lastUpdateCheckFilePath = constants.MakeMiniPath("last_update_check")
	githubClient            *github.Client
)

func MaybeUpdateFromGithub(output io.Writer) {
	MaybeUpdate(output, githubOwner, githubRepo, githubRepo, lastUpdateCheckFilePath)
}

func MaybeUpdate(output io.Writer, githubOwner, githubRepo, binaryName, lastUpdatePath string) {

	downloadBinary := binaryName + "-" + runtime.GOOS + "-" + runtime.GOARCH
	updateLinkPrefix := "https://github.com/" + githubOwner + "/" + githubRepo + "/releases/tag/" + version.VersionPrefix
	downloadLinkFormat := "https://github.com/" + githubOwner + "/" + githubRepo + "/releases/download/v%s/%s"

	if !shouldCheckURLVersion(lastUpdatePath) {
		return
	}
	latestVersion, err := getLatestVersionFromGitHub(githubOwner, githubRepo)
	if err != nil {
		glog.Errorln(err)
		return
	}
	localVersion, err := version.GetSemverVersion()
	if err != nil {
		glog.Errorln(err)
		return
	}
	if localVersion.Compare(latestVersion) < 0 {
		writeTimeToFile(lastUpdateCheckFilePath, time.Now().UTC())
		fmt.Fprintf(output, `There is a newer version of %s available (%s%s). Do you want to
automatically update now [y/N]? `,
			githubRepo, version.VersionPrefix, latestVersion)

		var confirm string
		fmt.Scanln(&confirm)

		if confirm == "y" {
			fmt.Printf("Updating to version %s\n", latestVersion)
			updateBinary(latestVersion, downloadBinary, updateLinkPrefix, downloadLinkFormat)
			return
		}

		fmt.Println("Skipping autoupdate")
	}
}

func shouldCheckURLVersion(filePath string) bool {
	if !viper.GetBool(config.WantUpdateNotification) {
		return false
	}
	lastUpdateTime := getTimeFromFileIfExists(filePath)
	if time.Since(lastUpdateTime).Hours() < viper.GetFloat64(config.ReminderWaitPeriodInHours) {
		return false
	}
	return true
}

func getLatestVersionFromGitHub(githubOwner, githubRepo string) (semver.Version, error) {
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
	var (
		release *github.RepositoryRelease
		resp    *github.Response
		err     error
	)
	release, resp, err = client.Repositories.GetLatestRelease(githubOwner, githubRepo)
	if err != nil {
		return semver.Version{}, err
	}
	defer resp.Body.Close()
	latestVersionString := release.TagName
	if latestVersionString != nil {
		return semver.Make(strings.TrimPrefix(*latestVersionString, "v"))

	}
	return semver.Version{}, fmt.Errorf("Cannot get release name")
}

func writeTimeToFile(path string, inputTime time.Time) error {
	err := ioutil.WriteFile(path, []byte(inputTime.Format(timeLayout)), 0644)
	if err != nil {
		return fmt.Errorf("Error writing current update time to file: %s", err)
	}
	return nil
}

func getTimeFromFileIfExists(path string) time.Time {
	lastUpdateCheckTime, err := ioutil.ReadFile(path)
	if err != nil {
		return time.Time{}
	}
	timeInFile, err := time.Parse(timeLayout, string(lastUpdateCheckTime))
	if err != nil {
		return time.Time{}
	}
	return timeInFile
}

func updateBinary(v semver.Version, downloadBinary, updateLinkPrefix, downloadLinkFormat string) {
	checksum, err := downloadChecksum(v, downloadBinary, downloadLinkFormat)
	if err != nil {
		glog.Errorf("Cannot download checksum: %s", err)
		os.Exit(1)
	}
	binary, err := http.Get(fmt.Sprintf(downloadLinkFormat, v, downloadBinary))
	if err != nil {
		glog.Errorf("Cannot download binary: %s", err)
		os.Exit(1)
	}
	defer binary.Body.Close()
	err = update.Apply(binary.Body, update.Options{
		Hash:     crypto.SHA256,
		Checksum: checksum,
	})
	if err != nil {
		glog.Errorf("Cannot apply binary update: %s", err)
		os.Exit(1)
	}

	env := os.Environ()
	args := os.Args
	currentBinary, err := osext.Executable()
	if err != nil {
		glog.Errorf("Cannot find current binary to exec: %s", err)
		os.Exit(1)
	}
	err = syscall.Exec(currentBinary, args, env)
	if err != nil {
		glog.Errorf("Failed to exec updated binary: %s", err)
		os.Exit(1)
	}
}

func downloadChecksum(v semver.Version, downloadBinary, downloadLinkFormat string) ([]byte, error) {
	u := fmt.Sprintf(downloadLinkFormat, v, downloadBinary+".sha256")
	checksumResp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer checksumResp.Body.Close()
	if checksumResp.StatusCode != 200 {
		return nil, fmt.Errorf("received %d", checksumResp.StatusCode)
	}
	b, err := ioutil.ReadAll(checksumResp.Body)
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(strings.TrimSpace(string(b)))
}
