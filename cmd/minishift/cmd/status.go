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
	"path/filepath"
	"strings"
	"text/template"

	"github.com/docker/go-units"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	cmdState "github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	openshiftVersion "github.com/minishift/minishift/pkg/minishift/openshift/version"
	"github.com/minishift/minishift/pkg/minishift/registration"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var statusFormat = `Minishift:  {{.MinishiftStatus}}
Profile:    {{.ProfileName}}
OpenShift:  {{.ClusterStatus}}
DiskUsage:  {{.DiskUsage}}
CacheUsage: {{.CacheUsage}} (used by oc binary, ISO or cached images)
`

var statusFormatWithRegistration = `Minishift:  {{.MinishiftStatus}}
Profile:    {{.ProfileName}}
OpenShift:  {{.ClusterStatus}}
DiskUsage:  {{.DiskUsage}}
CacheUsage: {{.CacheUsage}} (used by oc binary, ISO or cached images)
RHSM: 	    {{.Registration}}
`

type Status struct {
	MinishiftStatus string
	ProfileName     string
	ClusterStatus   string
	DiskUsage       string
	CacheUsage      string
}

type StatusWithRegistration struct {
	Status
	Registration string
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Gets the status of the local OpenShift cluster.",
	Long:  `Gets the status of the local OpenShift cluster.`,
	Run:   runStatus,
}

func runStatus(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(cmdState.InstanceDirs.Home, cmdState.InstanceDirs.Certs)
	defer api.Close()

	host, err := api.Load(constants.MachineName)
	if err != nil {
		s, err := cluster.GetHostStatus(api, constants.MachineName)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error getting cluster status: %s", err.Error()))
		}
		atexit.ExitWithMessage(0, s)
	}
	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}

	openshiftStatus := "Stopped"
	diskUsage := "Unknown"
	cacheUsage := "Unknown"
	profileName := constants.ProfileName

	vmStatus, err := cluster.GetHostStatus(api, constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting cluster status: %s", err.Error()))
	}

	var supportsRegistration bool
	rhelRegistration := "Not Registered"

	if vmStatus == state.Running.String() {
		openshiftVersion, err := openshiftVersion.GetOpenshiftVersion(sshCommander)
		if err == nil {
			openshiftStatus = fmt.Sprintf("Running (%s)", strings.Split(openshiftVersion, "\n")[0])
		}

		diskSize, diskUse, mountpoint := getDiskUsage(host.Driver, StorageDisk)
		if host.Driver.DriverName() == "generic" {
			diskSize, diskUse, mountpoint = getDiskUsage(host.Driver, StorageDiskForGeneric)
		}
		diskUsage = fmt.Sprintf("%s of %s (Mounted On: %s)", diskUse, diskSize, mountpoint)

		_, supportsRegistration, _ = registration.DetectRegistrator(sshCommander)
		if supportsRegistration {
			redHatRegistrator := registration.NewRedHatRegistrator(sshCommander)
			if registered, err := redHatRegistrator.IsRegistered(); registered && err == nil {
				rhelRegistration = "Registered"
			}
		}
	}

	cacheDir := filepath.Join(constants.GetMinishiftHomeDir(), "cache")
	var size int64
	err = filepath.Walk(cacheDir, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error finding size of cache: %s", err.Error()))
	}

	cacheUsage = units.HumanSize(float64(size))
	if supportsRegistration {
		status := StatusWithRegistration{Status{vmStatus, profileName, openshiftStatus, diskUsage, cacheUsage}, rhelRegistration}
		printStatus(status, statusFormatWithRegistration)
	} else {
		status := Status{vmStatus, profileName, openshiftStatus, diskUsage, cacheUsage}
		printStatus(status, statusFormat)
	}
}

func printStatus(status interface{}, statusFormat string) {
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
