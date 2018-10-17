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
	"github.com/minishift/minishift/cmd/minishift/cmd/addon"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/progressdots"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"fmt"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/provision"
	confCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minishift/clusterup"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the OpenShift cluster.",
	Long:  "Starts the OpenShift cluster and maintains the state of the Minishift VM.",
	Run:   runStart,
}

var clusterUpFlagSet *flag.FlagSet

func init() {
	clusterUpFlagSet = cmdUtil.InitClusterUpFlags("start")
	startCmd.Flags().AddFlagSet(clusterUpFlagSet)
	if enableExperimental, _ := cmdUtil.GetBoolEnv(minishiftConstants.MinishiftEnableExperimental); enableExperimental {
		OpenShiftCmd.AddCommand(startCmd)
	}
}

func runStart(cmd *cobra.Command, args []string) {
	api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer api.Close()

	host, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		atexit.ExitWithMessage(1, nonExistentMachineError)
	}

	sshCommander := provision.GenericSSHCommander{Driver: host.Driver}
	dockerbridgeSubnet, err := sshCommander.SSHCommand(minishiftConstants.DockerbridgeSubnetCmd)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	if minishiftConfig.InstanceStateConfig == nil {
		minishiftConfig.InstanceStateConfig, _ = minishiftConfig.NewInstanceStateConfig(minishiftConstants.GetInstanceStateConfigPath())
	}

	ocPath := minishiftConfig.InstanceStateConfig.OcPath
	openshiftVersion := minishiftConfig.InstanceStateConfig.OpenshiftVersion

	clusterUpConfig := &clusterup.ClusterUpConfig{
		OpenShiftVersion:     openshiftVersion,
		MachineName:          constants.MachineName,
		Ip:                   ip,
		Port:                 constants.APIServerPort,
		RoutingSuffix:        confCmd.GetDefaultRoutingSuffix(ip),
		User:                 minishiftConstants.DefaultUser,
		Project:              minishiftConstants.DefaultProject,
		KubeConfigPath:       constants.KubeConfigPath,
		OcPath:               ocPath,
		AddonEnv:             viper.GetStringSlice(confCmd.AddonEnv.Name),
		PublicHostname:       confCmd.GetDefaultPublicHostName(ip),
		SSHCommander:         sshCommander,
		OcBinaryPathInsideVM: fmt.Sprintf("%s/oc", minishiftConstants.OcPathInsideVM),
		SshUser:              sshCommander.Driver.GetSSHUsername(),
	}

	clusterUpParams := cmdUtil.DetermineClusterUpParameters(clusterUpConfig, dockerbridgeSubnet, clusterUpFlagSet)

	progressDots := progressdots.New()
	progressDots.Start()
	out, err := clusterup.ClusterUp(clusterUpConfig, clusterUpParams)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error during 'cluster up' execution: %v", err))
	}
	progressDots.Stop()
	fmt.Printf("\n%s\n", out)

	// postClusterUp performs configuration action which only need to be run after an initial provision of OpenShift.
	// On subsequent VM restarts these actions can be skipped.
	err = clusterup.PostClusterUp(clusterUpConfig, sshCommander, addon.GetAddOnManager(), &util.RealRunner{})
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error during post cluster up configuration: %v", err))
	}

	fmt.Println("OpenShift Started successfully")
}
