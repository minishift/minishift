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
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/pborman/uuid"
	"strings"
)

const OPENSHIFT_CONTAINER_NAME = "origin"
const OPENSHIFT_EXEC = "/usr/bin/openshift"

type OpenShiftPatchTarget struct {
	target              string
	containerConfigFile string
	localConfigFile     string
	ip                  string
}

var (
	MASTER = OpenShiftPatchTarget{
		"master",
		"/var/lib/origin/openshift.local.config/master/master-config.yaml",
		"/var/lib/minishift/openshift.local.config/master/master-config.yaml",
		"",
	}

	NODE = OpenShiftPatchTarget{
		"node",
		"/var/lib/origin/openshift.local.config/node-%s/node-config.yaml",
		"/var/lib/minishift/openshift.local.config/node-%s/node-config.yaml",
		"",
	}

	UNKNOWN = OpenShiftPatchTarget{
		"unkown",
		"",
		"",
		"",
	}
)

func (t *OpenShiftPatchTarget) SetIp(ip string) {
	t.ip = ip
}

func (t *OpenShiftPatchTarget) containerConfigFilePath() (string, error) {
	if t.target == "node" {
		if t.ip == "" {
			return "", errors.New("Node IP was not set")
		}
		return fmt.Sprintf(t.containerConfigFile, t.ip), nil
	} else {
		return t.containerConfigFile, nil
	}
}

func (t *OpenShiftPatchTarget) localConfigFilePath() (string, error) {
	if t.target == "node" {
		if t.ip == "" {
			return "", errors.New("Node IP was not set")
		}
		return fmt.Sprintf(t.localConfigFile, t.ip), nil
	} else {
		return t.localConfigFile, nil
	}
}

func RestartOpenShift(commander docker.DockerCommander) (bool, error) {
	fmt.Println("Restarting OpenShift")
	ok, err := commander.Restart(OPENSHIFT_CONTAINER_NAME)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func Patch(target OpenShiftPatchTarget, patch string, commander docker.DockerCommander) (bool, error) {
	fmt.Println(fmt.Sprintf("Patching OpenShift configuration %s with %s", target.containerConfigFile, patch))

	patchId, err := backUpConfig(target, commander)
	if err != nil {
		return false, err
	}

	containerConfigPath, err := target.containerConfigFilePath()
	if err != nil {
		return false, err
	}

	patchCommand := fmt.Sprintf("ex config patch %s --patch='%s'", containerConfigPath, patch)
	result, err := commander.Exec("-t", OPENSHIFT_CONTAINER_NAME, OPENSHIFT_EXEC, patchCommand)
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

func ViewConfig(target OpenShiftPatchTarget, commander docker.DockerCommander) (string, error) {
	path, err := target.containerConfigFilePath()
	if err != nil {
		return "", err
	}

	result, err := commander.Exec("-t", OPENSHIFT_CONTAINER_NAME, "cat", path)
	if err != nil {
		return "", err
	}
	return result, nil
}

func rollback(target OpenShiftPatchTarget, commander docker.DockerCommander, patchId string) {
	fmt.Println("Unable to restart OpenShift after patchting it. Rolling back.")
	restoreConfig(target, patchId, commander)
	commander.Restart(OPENSHIFT_CONTAINER_NAME)
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

	// TODO Figure out where the (meta) charcater 'M' comes from at the end of the line (HF)
	writeCommand := fmt.Sprintf("sudo echo '%s' | sed 's/.$//' | sudo tee %s > /dev/null", config, path)

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

	backupDeleteCommand := fmt.Sprintf("sudo rm %s-%s", path, id)

	_, err = commander.LocalExec(backupDeleteCommand)
	if err != nil {
		return err
	}

	return nil
}
