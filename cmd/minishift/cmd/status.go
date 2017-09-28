/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	openshiftVersion "github.com/minishift/minishift/pkg/minishift/openshift/version"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var statusFormat = `Minishift:  {{.MinishiftStatus}}
Profile:    {{.ProfileName}}
OpenShift:  {{.ClusterStatus}}
DiskUsage:  {{.DiskUsage}}
`

type Status struct {
	MinishiftStatus string
	ProfileName     string
	ClusterStatus   string
	DiskUsage       string
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Gets the status of the local OpenShift cluster.",
	Long:  `Gets the status of the local OpenShift cluster.`,
	Run:   runStatus,
}

func runStatus(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	host, err := api.Load(constants.MachineName)
	if err != nil {
		s, err := cluster.GetHostStatus(api)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error getting cluster status: %s", err.Error()))
		}
		atexit.ExitWithMessage(0, s)
	}

	openshiftStatus := "Stopped"
	diskUsage := "Unknown"
	profileName := profileActions.GetActiveProfile()

	vmStatus, err := cluster.GetHostStatus(api)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting cluster status: %s", err.Error()))
	}

	if vmStatus == state.Running.String() {
		openshiftVersion, err := openshiftVersion.GetOpenshiftVersion(host)
		if err == nil {
			openshiftStatus = fmt.Sprintf("Running (%s)", strings.Split(openshiftVersion, "\n")[0])
		}

		diskSize, diskUse := getDiskUsage(host.Driver, StorageDisk)
		diskUsage = fmt.Sprintf("%s of %s", diskUse, diskSize)
	}

	status := Status{vmStatus, profileName, openshiftStatus, diskUsage}

	tmpl, err := template.New("status").Parse(statusFormat)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating status template: %s", err.Error()))
	}
	err = tmpl.Execute(os.Stdout, status)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error executing status template:: %s", err.Error()))
	}
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
