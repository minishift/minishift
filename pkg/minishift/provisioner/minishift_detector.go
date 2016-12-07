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

package provisioner

import (
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/provision"
)

type MinishiftProvisionerDetector struct {
	Delegate provision.Detector
}

func (detector *MinishiftProvisionerDetector) DetectProvisioner(driver drivers.Driver) (provision.Provisioner, error) {
	log.Info("Waiting for SSH to be available...")
	if err := drivers.WaitForSSH(driver); err != nil {
		return nil, err
	}
	osReleaseOut, err := drivers.RunSSHCommandFromDriver(driver, "cat /etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("Error getting SSH command: %s", err)
	}

	log.Info("Detecting the provisioner...")
	osReleaseInfo, err := provision.NewOsRelease([]byte(osReleaseOut))
	if err != nil {
		return nil, fmt.Errorf("Error parsing /etc/os-release file: %s", err)
	}

	if detector.isMinishiftIso(osReleaseInfo) {
		provisioner := NewMinishiftProvisioner("minishift", driver)
		provisioner.SetOsReleaseInfo(osReleaseInfo)
		return provisioner, nil
	} else {
		return detector.Delegate.DetectProvisioner(driver)
	}
	return nil, provision.ErrDetectionFailed
}

func (detector *MinishiftProvisionerDetector) isMinishiftIso(osReleaseInfo *provision.OsRelease) bool {
	if osReleaseInfo.Variant == "minishift" {
		return true
	}
	return false
}
