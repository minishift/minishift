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

package minishift

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	pb "gopkg.in/cheggaaa/pb.v1"

	"github.com/minishift/minishift/pkg/util/github"
	"github.com/minishift/minishift/pkg/version"
)

const OC_BINARY_NAME = "oc"

// Oc is a struct with methods designed for dealing with the oc binary
type Oc struct {
	OpenShiftVersion  string
	MinishiftCacheDir string
}

func (oc *Oc) EnsureIsCached() error {
	if !oc.isCached() {
		err := oc.cacheOc()
		if err != nil {
			return err
		}

	}
	return nil
}

func (oc *Oc) GetCacheFilepath() string {
	return filepath.Join(oc.MinishiftCacheDir, OC_BINARY_NAME, oc.OpenShiftVersion)
}

func (oc *Oc) isCached() bool {
	if _, err := os.Stat(filepath.Join(oc.GetCacheFilepath(), OC_BINARY_NAME)); os.IsNotExist(err) {
		return false
	}
	return true
}

// cacheOc downloads and caches the oc binary into the minishift directory
func (oc *Oc) cacheOc() error {
	urlObj, err := url.Parse(oc.OpenShiftVersion)
	if err != nil {
		return errors.Wrap(err, "Error parsing --openshift-version url")
	}
	if urlObj.Scheme == "file" {
		f, err := os.Open(strings.Replace(urlObj.Path, "/", string(filepath.Separator), -1))
		if err != nil {
			return errors.Wrapf(err, "Error opening specified OpenShift file %s", oc.OpenShiftVersion)
		}
		defer f.Close()
		stat, err := f.Stat()
		_, err = f.Stat()
		if err != nil {
			return errors.Wrapf(err, "Error opening specified OpenShift file %s", urlObj.Path)
		}
		return oc.updateOpenShiftFromReader(f, stat.Size())
	} else {
		if _, err := semver.Make(strings.TrimPrefix(oc.OpenShiftVersion, version.VersionPrefix)); err != nil {
			resp, err := http.Get(oc.OpenShiftVersion)
			if err != nil {
				return errors.Wrapf(err, "Error downloading OpenShift from %s", oc.OpenShiftVersion)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return errors.Errorf("Error downloading OpenShift from %s: received status code %d", oc.OpenShiftVersion, resp.StatusCode)
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

			return oc.updateOpenShiftFromReader(openshiftFile, resp.ContentLength)
		} else {
			return oc.updateFromRelease()
		}
	}
}

func (oc *Oc) updateFromRelease() error {
	if !oc.isCached() {
		if err := github.DownloadOpenShiftReleaseBinary(github.OC, oc.determineOS(), oc.OpenShiftVersion, oc.GetCacheFilepath()); err != nil {
			return errors.Wrapf(err, "Error attempting to download and cache %s", github.OC.String())
		}
	}
	return nil
}

func (oc *Oc) determineOS() github.OS {
	switch runtime.GOOS {
	case "windows":
		return github.WINDOWS
	case "darwin":
		return github.DARWIN
	case "linux":
		return github.LINUX
	}
	panic("Unexpected OS type")
}

func (oc *Oc) updateOpenShiftFromReader(r io.Reader, size int64) error {
	// TODO needs to copy to minishift path
	//if err := sshutil.Transfer(r, size, "/usr/local/bin",
	//	"openshift", "0777", client); err != nil {
	//	return errors.Wrap(err, "Error transferring file to VM via ssh")
	//}
	return nil
}
