/*
Copyright (C) 2018 Red Hat, Inc.

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
	"os"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/clusterup"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	openshiftVersion "github.com/minishift/minishift/pkg/minishift/openshift/version"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	minishiftStrings "github.com/minishift/minishift/pkg/util/strings"
	"github.com/spf13/cobra"
)

const (
	nonSpecifiedComponentError = "You need to specify a component name (use 'minishift openshift component list' to find available components)"
	nonValidComponentError     = "You have specified a non-valid component name, use 'minishift openshift component list' to find valid components"
)

// version command represent current running openshift version and available one.
var componentAddCmd = &cobra.Command{
	Use:   "add [component-name]",
	Short: "Add component to an OpenShift cluster (Works only with OpenShift version >= 3.10.x)",
	Long:  "Add component to an OpenShift cluster (Works only with OpenShift version >= 3.10.x)",
	Run:   runComponentAdd,
}

var (
	component string
)

func runComponentAdd(cmd *cobra.Command, args []string) {
	if len(args) <= 0 {
		atexit.ExitWithMessage(1, nonSpecifiedComponentError)
	}

	component = args[0]

	if !minishiftStrings.Contains(minishiftConstants.ValidComponents, component) {
		atexit.ExitWithMessage(1, nonValidComponentError)
	}

	api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer api.Close()

	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		atexit.ExitWithMessage(1, nonExistentMachineError)
	}
	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}

	// Get proper OpenShift version
	minishiftConfig.InstanceStateConfig, err = minishiftConfig.NewInstanceStateConfig(minishiftConstants.GetInstanceStateConfigPath())
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error reading config for VM: %s", err.Error()))
	}

	requestedOpenShiftVersion := minishiftConfig.InstanceStateConfig.OpenshiftVersion
	valid, err := openshiftVersion.IsGreaterOrEqualToBaseVersion(requestedOpenShiftVersion, constants.RefactoredOcVersion)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}
	if !valid {
		atexit.ExitWithMessage(1, fmt.Sprintf("You are using %s but this feature only available for OpenShift >= 3.10.x", requestedOpenShiftVersion))
	}

	imageToUse := fmt.Sprintf("'%s:%s'", minishiftConstants.ImageNameForClusterUpImageFlag, requestedOpenShiftVersion)

	baseDirectory := minishiftConstants.BaseDirInsideInstance
	ocPathInsideVM := fmt.Sprintf("%s/oc", minishiftConstants.OcPathInsideVM)

	out, err := clusterup.AddComponent(sshCommander, ocPathInsideVM, baseDirectory, component, imageToUse)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}
	fmt.Fprint(os.Stdout, out)

}

func init() {
	componentCmd.AddCommand(componentAddCmd)
}
