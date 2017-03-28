/*
Copyright (C) 2017 Red Hat, Inc.

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

package hostfolder

import (
	"fmt"
	"github.com/golang/glog"
	"os"

	"github.com/docker/machine/libmachine"
	"github.com/minishift/minishift/pkg/minikube/constants"
	hostfolderActions "github.com/minishift/minishift/pkg/minishift/hostfolder"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var (
	mountAll bool
)

// HostfolderCmd represents the resource command for host folders
var HostfolderCmd = &cobra.Command{
	Use:   "hostfolder SUBCOMMAND [flags]",
	Short: "Manage and control host folders for use by the OpenShift cluster.",
	Long:  `Manage and control host folders for use by the OpenShift cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var hostfolderMountCmd = &cobra.Command{
	Use:   "mount HOSTFOLDER_NAME",
	Short: "Mount a host folder to the running cluster",
	Long:  `Mount a host folder to the running cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()
		host, _ := api.Load(constants.MachineName)

		var err error = nil
		if mountAll {
			err = hostfolderActions.MountHostfolders(host.Driver)
		} else {
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, "usage: minishift hostfolder mount [HOSTFOLDER_NAME|--all]")
				atexit.Exit(1)
			}
			err = hostfolderActions.Mount(host.Driver, args[0])
		}

		if err != nil {
			glog.Errorln(err)
			atexit.Exit(1)
		}

	},
}

// TODO: refactor; lot of repetition with previous command
var hostfolderUmountCmd = &cobra.Command{
	Use:   "umount HOSTFOLDER_NAME",
	Short: "Unmount a host folder from the running cluster",
	Long:  `Unmount a host folder from the running cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "usage: minishift hostfolder umount HOSTFOLDER_NAME")
			atexit.Exit(1)
		}

		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()
		host, _ := api.Load(constants.MachineName)

		err := hostfolderActions.Umount(host.Driver, args[0])
		if err != nil {
			glog.Errorln(err)
			atexit.Exit(1)
		}
	},
}

var hostfolderListCmd = &cobra.Command{
	Use:   "list",
	Short: "List an overview of defined host folders",
	Long:  `List an overview of defined host folders that can be mounted to a running cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()
		host, _ := api.Load(constants.MachineName)

		hostfolderActions.List(host.Driver)
	},
}

var hostfolderAddCmd = &cobra.Command{
	Use:   "add HOSTFOLDER_NAME",
	Short: "Add a host folder definition",
	Long:  `Add a host folder definition that can be mounted to a running cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "usage: minishift hostfolder add HOSTFOLDER_NAME")
			atexit.Exit(1)
		}

		hostfolderActions.Add(args[0])
	},
}

var hostfolderSetupUsersCmd = &cobra.Command{
	Use:   "setup-users",
	Short: "Setup credentials for the Users share on a Windows host",
	Long:  `Setup credentials for the Users share on a Windows host`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		hostfolderActions.SetupUsers()
	},
}

var hostfolderRemoveCmd = &cobra.Command{
	Use:   "remove HOSTFOLDER_NAME",
	Short: "Remove a host folder definition",
	Long:  `Remove a host folder definition`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "usage: minishift hostfolder remove HOSTFOLDER_NAME")
			atexit.Exit(1)
		}

		hostfolderActions.Remove(args[0])
	},
}

func init() {
	// parent set in root.go (RootCmd)
	//HostfolderCmd.PersistentFlags().Bool(hostfolderActions.AutoMountHostfolders, false, "Auto-mount hostfolders on startup")
	HostfolderCmd.AddCommand(hostfolderMountCmd)
	hostfolderMountCmd.Flags().BoolVarP(&mountAll, "all", "a", false, "Mount all defined host folders to the cluster instance.")
	HostfolderCmd.AddCommand(hostfolderUmountCmd)
	HostfolderCmd.AddCommand(hostfolderListCmd)
	HostfolderCmd.AddCommand(hostfolderAddCmd)
	HostfolderCmd.AddCommand(hostfolderRemoveCmd)
	HostfolderCmd.AddCommand(hostfolderSetupUsersCmd)
}
