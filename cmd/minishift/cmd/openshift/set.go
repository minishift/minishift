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
	"github.com/spf13/cobra"

	"github.com/docker/machine/libmachine"

	"encoding/json"
	"fmt"

	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/minishift/minishift/pkg/minishift/openshift"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

const (
	targetFlag = "target"
	patchFlag  = "patch"

	unknownPatchTargetError = "Unkown patch target. Only 'master' and 'node' are supported."
	emptyPatchError         = "You need to specify a patch using --patch."
	invalidJSONError        = "The specified patch need to be valid JSON."
)

var (
	target string
	patch  string
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Updates the specified OpenShift configuration resource with the specified patch.",
	Long:  "Updates the specified OpenShift configuration resource with the specified patch. The patch needs to be in JSON format.",
	Run:   runPatch,
}

func init() {
	setCmd.Flags().StringVar(&target, targetFlag, "master", "Target configuration to patch. Either 'master' or 'node'.")
	setCmd.Flags().StringVar(&patch, patchFlag, "", "The patch to apply")
	configCmd.AddCommand(setCmd)
}

func runPatch(cmd *cobra.Command, args []string) {
	patchTarget := determineTarget(target)
	if patchTarget == openshift.UNKNOWN {
		atexit.ExitWithMessage(1, unknownPatchTargetError)
	}

	validatePatch(patch)

	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		atexit.ExitWithMessage(1, nonExistentMachineError)
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		atexit.ExitWithMessage(1, unableToRetrieveIpError)
	}
	patchTarget.SetIp(ip)

	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}
	dockerCommander := docker.NewVmDockerCommander(sshCommander)

	_, err = openshift.Patch(patchTarget, patch, dockerCommander)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error patching OpenShift configuration: %s", err.Error()))
	}
}

func determineTarget(target string) openshift.OpenShiftPatchTarget {
	switch target {
	case "master":
		return openshift.MASTER
	case "node":
		return openshift.NODE
	default:
		return openshift.UNKNOWN
	}
}

func validatePatch(patch string) {
	if len(patch) == 0 {
		atexit.ExitWithMessage(1, emptyPatchError)
	}

	if !isJSON(patch) {
		atexit.ExitWithMessage(1, invalidJSONError)
	}

}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil

}
