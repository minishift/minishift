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

package openshift

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/minikube/constants"
	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/minishift/docker"
	openshiftVersionCheck "github.com/minishift/minishift/pkg/minishift/openshift/version"
	"github.com/pborman/uuid"
)

type OpenShiftPatchTarget struct {
	target              string
	containerConfigFile string
	localConfigFile     string
}

var (
	MASTER = OpenShiftPatchTarget{
		"master",
		getContainerConfigFile("master"),
		getLocalConfigFile("master"),
	}

	NODE = OpenShiftPatchTarget{
		"node",
		getContainerConfigFile("node"),
		getLocalConfigFile("node"),
	}

	UNKNOWN = OpenShiftPatchTarget{
		"unkown",
		"",
		"",
	}
)

func (t *OpenShiftPatchTarget) containerConfigFilePath() (string, error) {
	if t.target == "node" {
		return t.containerConfigFile, nil
	} else {
		return t.containerConfigFile, nil
	}
}

func (t *OpenShiftPatchTarget) localConfigFilePath() (string, error) {
	if t.target == "node" {
		return t.localConfigFile, nil
	} else {
		return t.localConfigFile, nil
	}
}

func RestartOpenShift(commander docker.DockerCommander) (bool, error) {
	var (
		ok  bool
		err error
	)
	openshiftVersion := getOpenshiftVersion()
	valid, _ := openshiftVersionCheck.IsGreaterOrEqualToBaseVersion(openshiftVersion, constants.RefactoredOcVersion)
	if valid {
		containerID, err := commander.GetID(minishiftConstants.OpenshiftApiContainerName)
		if err != nil {
			return false, err
		}
		ok, err = commander.Stop(containerID)
		if err != nil {
			return false, err
		}
	}
	ok, err = commander.Restart(minishiftConstants.OpenshiftContainerName)
	if err != nil {
		return false, err
	}

	return ok, err
}

func Patch(target OpenShiftPatchTarget, patch string, commander docker.DockerCommander, openshiftVersion string) (bool, error) {
	fmt.Println(fmt.Sprintf("Patching OpenShift configuration '%s' with '%s'", target.containerConfigFile, patch))

	patchId, err := backUpConfig(target, commander)
	if err != nil {
		return false, err
	}

	localConfigPath, err := target.localConfigFilePath()
	if err != nil {
		return false, err
	}

	patchCommand := fmt.Sprintf("ex config patch %s --patch='%s'", localConfigPath, patch)
	cmd := fmt.Sprintf("%s/oc %s", minishiftConstants.OcPathInsideVM, patchCommand)

	result, err := commander.LocalExec(cmd)
	if err != nil {
		glog.Error("Creating patched configuration failed. Not applying the changes.", err)
		return false, nil

	}

	// Tweak the result configuration, we need to escape single quotes
	result = strings.Replace(result, "'", "'\\''", -1)
	writeConfig(target, result, commander)
	if err != nil {
		rollback(target, commander, patchId)
		return false, nil
	}

	_, err = RestartOpenShift(commander)
	if err != nil {
		rollback(target, commander, patchId)
		return false, nil
	}

	deleteBackup(target, patchId, commander)

	return true, nil
}

// IsRunning checks whether the origin container is in running state.
// This method returns true if the origin container is running, false otherwise
func IsRunning(commander docker.DockerCommander) bool {
	status, err := commander.Status(minishiftConstants.OpenshiftContainerName)
	if err != nil || status != "running" {
		return false
	}

	return true
}

func ViewConfig(target OpenShiftPatchTarget, commander docker.DockerCommander) (string, error) {
	var (
		result string
		err    error
	)
	path, err := target.containerConfigFilePath()
	if err != nil {
		return "", err
	}
	openshiftVersion := getOpenshiftVersion()
	valid, _ := openshiftVersionCheck.IsGreaterOrEqualToBaseVersion(openshiftVersion, constants.RefactoredOcVersion)
	if !valid || target.target == "node" {
		result, err = commander.Exec("-t", minishiftConstants.OpenshiftContainerName, "cat", path)
		if err != nil {
			return "", err
		}
	} else {
		containerID, err := commander.GetID(minishiftConstants.OpenshiftApiContainerName)
		if err != nil {
			return "", err
		}
		result, err = commander.Exec("-t", containerID, "cat", path)
		if err != nil {
			return "", err
		}
	}

	return result, nil
}

func rollback(target OpenShiftPatchTarget, commander docker.DockerCommander, patchId string) {
	fmt.Println("Unable to restart OpenShift after patchting it. Rolling back.")
	restoreConfig(target, patchId, commander)
	commander.Restart(minishiftConstants.OpenshiftContainerName)
	deleteBackup(target, patchId, commander)
}

func backUpConfig(target OpenShiftPatchTarget, commander docker.DockerCommander) (string, error) {
	path, err := target.localConfigFilePath()
	if err != nil {
		return "", err
	}

	id := uuid.New()
	backupCommand := fmt.Sprintf("sudo cp %s %s-%s", path, path, id)

	_, err = commander.LocalExec(backupCommand)
	if err != nil {
		return "", err
	}

	return id, nil
}

func restoreConfig(target OpenShiftPatchTarget, id string, commander docker.DockerCommander) error {
	path, err := target.localConfigFilePath()
	if err != nil {
		return err
	}

	restoreCommand := fmt.Sprintf("sudo cp %s-%s %s", path, id, path)

	_, err = commander.LocalExec(restoreCommand)
	if err != nil {
		return err
	}

	return nil
}

func writeConfig(target OpenShiftPatchTarget, config string, commander docker.DockerCommander) error {
	path, err := target.localConfigFilePath()
	if err != nil {
		return err
	}

	writeCommand := fmt.Sprintf("sudo echo '%s' | sudo tee %s > /dev/null", config, path)

	_, err = commander.LocalExec(writeCommand)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func deleteBackup(target OpenShiftPatchTarget, id string, commander docker.DockerCommander) error {
	path, err := target.localConfigFilePath()
	if err != nil {
		return err
	}

	backupDeleteCommand := fmt.Sprintf("sudo rm -f %s-%s", path, id)

	_, err = commander.LocalExec(backupDeleteCommand)
	if err != nil {
		return err
	}

	return nil
}

func getLocalConfigFile(target string) string {
	openshiftVersion := getOpenshiftVersion()
	valid, _ := openshiftVersionCheck.IsGreaterOrEqualToBaseVersion(openshiftVersion, constants.RefactoredOcVersion)
	switch target {
	case "master":
		if !valid {
			return "/var/lib/minishift/openshift.local.config/master/master-config.yaml"
		}
		return "/var/lib/minishift/base/openshift-apiserver/master-config.yaml"
	case "node":
		if !valid {
			return "/var/lib/minishift/openshift.local.config/node-localhost/node-config.yaml"
		}
		return "/var/lib/minishift/base/node/node-config.yaml"
	}
	return ""
}

func getContainerConfigFile(target string) string {
	openshiftVersion := getOpenshiftVersion()
	valid, _ := openshiftVersionCheck.IsGreaterOrEqualToBaseVersion(openshiftVersion, constants.RefactoredOcVersion)
	switch target {
	case "master":
		if !valid {
			return "/var/lib/origin/openshift.local.config/master/master-config.yaml"
		}
		return "/etc/origin/master/master-config.yaml"
	case "node":
		if !valid {
			return "/var/lib/origin/openshift.local.config/node-localhost/node-config.yaml"
		}
		return "/var/lib/origin/openshift.local.config/node/node-config.yaml"
	}
	return ""
}

func getOpenshiftVersion() string {
	var openshiftVersion string

	instanceState.InstanceStateConfig, _ = instanceState.NewInstanceStateConfig(minishiftConstants.GetInstanceStateConfigPath())

	if instanceState.InstanceStateConfig != nil {
		openshiftVersion = instanceState.InstanceStateConfig.OpenshiftVersion
	}
	return openshiftVersion
}
