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
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/pborman/uuid"
)

type OpenShiftPatchTarget struct {
	target              string
	containerConfigFile string
	LocalConfigFile     string
}

func GetOpenShiftPatchTarget(target string) OpenShiftPatchTarget {
	switch target {
	case "master":
		return OpenShiftPatchTarget{
			"master",
			getContainerConfigFile("master"),
			getLocalConfigFile("master"),
		}
	case "node":
		return OpenShiftPatchTarget{
			"node",
			getContainerConfigFile("node"),
			getLocalConfigFile("node"),
		}
	case "kube":
		return OpenShiftPatchTarget{
			"kube",
			getContainerConfigFile("kube"),
			getLocalConfigFile("kube"),
		}
	default:
		return OpenShiftPatchTarget{
			"unkown",
			"",
			"",
		}
	}
}

func (t *OpenShiftPatchTarget) containerConfigFilePath() string {
	return t.containerConfigFile
}

func (t *OpenShiftPatchTarget) localConfigFilePath() string {
	return t.LocalConfigFile
}

func RestartOpenShift(commander docker.DockerCommander) (bool, error) {
	var (
		ok  bool
		err error
	)
	containerID, err := commander.GetID(minishiftConstants.OpenshiftApiContainerLabel)
	if err != nil {
		return false, err
	}
	ok, err = commander.Stop(containerID)
	if err != nil {
		return false, err
	}
	containerID, err = commander.GetID(minishiftConstants.KubernetesApiContainerLabel)
	if err != nil {
		return false, err
	}
	ok, err = commander.Stop(containerID)
	if err != nil {
		return false, err
	}
	ok, err = commander.Restart(minishiftConstants.OpenshiftContainerName)
	if err != nil {
		return false, err
	}
	return ok, err
}

func Patch(target OpenShiftPatchTarget, patch string, commander docker.DockerCommander) (bool, error) {
	patchId, err := backUpConfig(target, commander)
	if err != nil {
		return false, err
	}

	localConfigPath := target.localConfigFilePath()

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
	path := target.containerConfigFilePath()
	if target.target == "node" {
		result, err := commander.Exec("-t", minishiftConstants.OpenshiftContainerName, "cat", path)
		if err != nil {
			return "", err
		}
		return result, nil
	}
	if target.target == "kube" {
		containerID, err := commander.GetID(minishiftConstants.KubernetesApiContainerLabel)
		if err != nil {
			return "", err
		}
		result, err := commander.Exec("-t", containerID, "cat", path)
		if err != nil {
			return "", err
		}
		return result, nil
	}
	containerID, err := commander.GetID(minishiftConstants.OpenshiftApiContainerLabel)
	if err != nil {
		return "", err
	}
	result, err := commander.Exec("-t", containerID, "cat", path)
	if err != nil {
		return "", err
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
	path := target.localConfigFilePath()

	id := uuid.New()
	backupCommand := fmt.Sprintf("sudo cp %s %s-%s", path, path, id)

	_, err := commander.LocalExec(backupCommand)
	if err != nil {
		return "", err
	}

	return id, nil
}

func restoreConfig(target OpenShiftPatchTarget, id string, commander docker.DockerCommander) error {
	path := target.localConfigFilePath()

	restoreCommand := fmt.Sprintf("sudo cp %s-%s %s", path, id, path)

	_, err := commander.LocalExec(restoreCommand)
	if err != nil {
		return err
	}

	return nil
}

func writeConfig(target OpenShiftPatchTarget, config string, commander docker.DockerCommander) error {
	path := target.localConfigFilePath()

	writeCommand := fmt.Sprintf("sudo echo '%s' | sudo tee %s > /dev/null", config, path)

	_, err := commander.LocalExec(writeCommand)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func deleteBackup(target OpenShiftPatchTarget, id string, commander docker.DockerCommander) error {
	path := target.localConfigFilePath()

	backupDeleteCommand := fmt.Sprintf("sudo rm -f %s-%s", path, id)

	_, err := commander.LocalExec(backupDeleteCommand)
	if err != nil {
		return err
	}

	return nil
}

func getLocalConfigFile(target string) string {
	switch target {
	case "master":
		return "/var/lib/minishift/base/openshift-apiserver/master-config.yaml"
	case "node":
		return "/var/lib/minishift/base/node/node-config.yaml"
	case "kube":
		return "/var/lib/minishift/base/kube-apiserver/master-config.yaml"
	}
	return ""
}

func getContainerConfigFile(target string) string {
	switch target {
	case "master":
		return "/etc/origin/master/master-config.yaml"
	case "node":
		return "/var/lib/origin/openshift.local.config/node/node-config.yaml"
	case "kube":
		return "/etc/origin/master/master-config.yaml"
	}
	return ""
}
