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

package cache

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	pb "gopkg.in/cheggaaa/pb.v1"

	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minikube/sshutil"
	"github.com/minishift/minishift/pkg/util/github"
	minishiftos "github.com/minishift/minishift/pkg/util/os"
	"github.com/minishift/minishift/pkg/version"
)

const (
	fileScheme = "file"
)

// openshiftCacher is a struct with methods designed for caching openshift
type openshiftCacher struct {
	config   cluster.MachineConfig
	ProxyUrl string
}

func (l *openshiftCacher) getOpenShiftCacheFilepath() string {
	return filepath.Join(constants.Minipath, "cache", "openshift",
		filepath.Base(url.QueryEscape("openshift-"+l.config.OpenShiftVersion)))
}

func (l *openshiftCacher) isOpenShiftCached() bool {
	if _, err := os.Stat(l.getOpenShiftCacheFilepath()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (l *openshiftCacher) updateOpenShiftFromURI(client *ssh.Client) error {
	urlObj, err := url.Parse(l.config.OpenShiftVersion)
	if err != nil {
		return errors.Wrap(err, "Error parsing --openshift-version url")
	}
	if urlObj.Scheme == fileScheme {
		f, err := os.Open(strings.Replace(urlObj.Path, "/", string(filepath.Separator), -1))
		if err != nil {
			return errors.Wrapf(err, "Error opening specified OpenShift file %s", l.config.OpenShiftVersion)
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil {
			return errors.Wrapf(err, "Error opening specified OpenShift file %s", urlObj.Path)
		}
		return l.updateOpenShiftFromReader(f, stat.Size(), client)
	} else {
		if _, err := semver.Make(strings.TrimPrefix(l.config.OpenShiftVersion, version.VersionPrefix)); err != nil {
			resp, err := http.Get(l.config.OpenShiftVersion)
			if err != nil {
				return errors.Wrapf(err, "Error downloading OpenShift from %s", l.config.OpenShiftVersion)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return errors.Errorf("Error downloading OpenShift from %s: received status code %d", l.config.OpenShiftVersion, resp.StatusCode)
			}
			openshiftFile := resp.Body

			if resp.ContentLength > 0 {
				bar := pb.New64(resp.ContentLength).SetUnits(pb.U_BYTES)
				bar.Start()
				openshiftFile = bar.NewProxyReader(openshiftFile)
				defer func() {
					<-time.After(bar.RefreshRate)
					fmt.Println()
				}()
			}

			return l.updateOpenShiftFromReader(openshiftFile, resp.ContentLength, client)
		} else {
			return l.updateOpenShiftFromRelease(client)
		}
	}
}

func (l *openshiftCacher) updateOpenShiftFromRelease(client *ssh.Client) error {
	if !l.isOpenShiftCached() {
		if err := github.DownloadOpenShiftReleaseBinary(github.OC, minishiftos.LINUX, l.config.OpenShiftVersion, l.getOpenShiftCacheFilepath(), l.ProxyUrl); err != nil {
			return errors.Wrap(err, "Error attempting to download and cache openshift")
		}
	}
	if err := l.transferCachedOpenShiftToVM(client); err != nil {
		return errors.Wrap(err, "Error transferring cached openshift to VM")
	}
	return nil
}

func (l *openshiftCacher) transferCachedOpenShiftToVM(client *ssh.Client) error {
	f, err := os.Open(l.getOpenShiftCacheFilepath())
	if err != nil {
		return errors.Wrap(err, "Error reading file: openshift cache filepath")
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return errors.Wrap(err, "Error reading file: openshift cache filepath")
	}

	if err = sshutil.Transfer(f, stat.Size(), "/usr/local/bin",
		"openshift", "0777", client); err != nil {
		return errors.Wrap(err, "Error transferring cached openshift to VM via ssh")
	}
	return nil
}

func (l *openshiftCacher) updateOpenShiftFromReader(r io.Reader, size int64, client *ssh.Client) error {
	if err := sshutil.Transfer(r, size, "/usr/local/bin",
		"openshift", "0777", client); err != nil {
		return errors.Wrap(err, "Error transferring file to VM via ssh")
	}
	return nil
}
