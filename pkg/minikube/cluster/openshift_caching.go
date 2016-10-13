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

package cluster

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"

	"github.com/jimmidyson/minishift/pkg/minikube/constants"
	"github.com/jimmidyson/minishift/pkg/minikube/sshutil"
	"github.com/jimmidyson/minishift/pkg/util/github"
)

func updateOpenShiftFromAsset(client *ssh.Client) error {
	contents, err := Asset("out/openshift")
	if err != nil {
		return errors.Wrap(err, "Error loading asset out/openshift")
	}
	if err := sshutil.Transfer(bytes.NewReader(contents), len(contents), "/usr/local/bin",
		"openshift", "0777", client); err != nil {
		return errors.Wrap(err, "Error transferring openshift via ssh")
	}
	return nil
}

// openshiftCacher is a struct with methods designed for caching openshift
type openshiftCacher struct {
	config MachineConfig
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
		return l.updateOpenShiftFromFile(client)
	} else {
		return l.updateOpenShiftFromURL(client)
	}
}

func (l *openshiftCacher) updateOpenShiftFromURL(client *ssh.Client) error {
	if !l.isOpenShiftCached() {
		if err := github.DownloadOpenShiftRelease(l.config.OpenShiftVersion, l.getOpenShiftCacheFilepath()); err != nil {
			return errors.Wrap(err, "Error attempting to download and cache openshift")
		}
	}
	if err := l.transferCachedOpenShiftToVM(client); err != nil {
		return errors.Wrap(err, "Error transferring cached openshift to VM")
	}
	return nil
}

func (l *openshiftCacher) transferCachedOpenShiftToVM(client *ssh.Client) error {
	contents, err := ioutil.ReadFile(l.getOpenShiftCacheFilepath())
	if err != nil {
		return errors.Wrap(err, "Error reading file: openshift cache filepath")
	}

	if err = sshutil.Transfer(bytes.NewReader(contents), len(contents), "/usr/local/bin",
		"openshift", "0777", client); err != nil {
		return errors.Wrap(err, "Error transferring cached openshift to VM via ssh")
	}
	return nil
}

func (l *openshiftCacher) updateOpenShiftFromFile(client *ssh.Client) error {
	path := strings.TrimPrefix(l.config.OpenShiftVersion, "file://")
	path = filepath.FromSlash(path)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "Error reading openshift file at %s", path)
	}
	if err := sshutil.Transfer(bytes.NewReader(contents), len(contents), "/usr/local/bin",
		"openshift", "0777", client); err != nil {
		return errors.Wrapf(err, "Error transferring specified openshift file at %s to VM via ssh", path)
	}
	return nil
}
