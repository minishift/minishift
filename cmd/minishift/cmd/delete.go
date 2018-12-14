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

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	registrationUtil "github.com/minishift/minishift/cmd/minishift/cmd/registration"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/minishift/oc"
	pkgUtil "github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	// deleteCmd represents the delete command
	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Deletes the Minishift VM.",
		Long:  `Deletes the Minishift VM, including the local OpenShift cluster and all associated files.`,
		Run:   runDelete,
	}

	forceFlag      bool
	clearCacheFlag bool
)

func runDelete(cmd *cobra.Command, args []string) {
	if clearCacheFlag {
		clearCache()
	}

	api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer api.Close()

	if !util.VMExists(api, constants.MachineName) {
		atexit.Exit(0)
	}

	util.ExitIfUndefined(api, constants.MachineName)

	host, err := api.Load(constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	if !forceFlag {
		hasConfirmed := pkgUtil.AskForConfirmation(fmt.Sprintf("You are deleting the Minishift VM: '%s'.", constants.MachineName))
		if !hasConfirmed {
			atexit.Exit(0)
		}
	}

	if host.Driver.DriverName() == "generic" {
		if err := util.OcClusterDown(host); err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
		if err := deleteExistingDirectory(host); err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
	}
	// Remove entries from global kube config
	clusterIP, _ := cluster.GetHostIP(api)
	err = cleanKubeConfig(clusterIP)
	if err != nil {
		fmt.Println("Unable to delete entries from kube config:", err)
	}

	// Unregistration, do not allow to be skipped
	registrationUtil.UnregisterHost(api, false, forceFlag)
	fmt.Println("Deleting the Minishift VM...")
	if err := cluster.DeleteHost(api); err != nil {
		handleFailedHostDeletion(err)
	}

	removeInstanceAndKubeConfig()

	fmt.Println("Minishift VM deleted.")
}

func clearCache() {
	if !forceFlag {
		hasConfirmed := pkgUtil.AskForConfirmation("This will delete the cache content for all profiles.")
		if !hasConfirmed {
			return
		}
	}
	cachePath := state.InstanceDirs.Cache
	err := os.RemoveAll(cachePath)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error deleting Minishift cache: %v", err))
	} else {
		fmt.Printf("Removed cache content at: '%s'\n", cachePath)
	}
}

// Remove the current cluster's entries from global kubeconfig file
func cleanKubeConfig(clusterIP string) error {
	kubeConfigPath, err := oc.GetGlobalKubeConfigPath()
	if err != nil {
		return err
	}
	kubeConfig, err := clientcmd.LoadFromFile(kubeConfigPath)
	if err != nil {
		return err
	}
	cleanKubeConfig, err := removeEntriesForCluster(clusterIP, kubeConfig)
	if err != nil {
		return err
	}
	return clientcmd.WriteToFile(*cleanKubeConfig, kubeConfigPath)
}

func removeEntriesForCluster(clusterIP string, kubeConfig *clientcmdapi.Config) (*clientcmdapi.Config, error) {
	if kubeConfig == nil {
		return nil, fmt.Errorf("Empty kubeconfig.")
	}
	clusterName := fmt.Sprintf("%s:%s", strings.Replace(clusterIP, ".", "-", -1), "8443")
	cleanKubeConfig := clientcmdapi.NewConfig()

	fmt.Printf("Removing entries from kubeconfig for cluster: %s\n", clusterName)
	for cName, c := range kubeConfig.Clusters {
		if cName == clusterName {
			continue
		}
		cleanKubeConfig.Clusters[cName] = c
	}
	for ctxName, ctx := range kubeConfig.Contexts {
		if ctx.Cluster == clusterName {
			continue
		}
		cleanKubeConfig.Contexts[ctxName] = ctx
	}
	for aName, a := range kubeConfig.AuthInfos {
		if strings.Contains(aName, clusterName) {
			continue
		}
		cleanKubeConfig.AuthInfos[aName] = a
	}
	// restore stuff that minishift doesnot touch
	cleanKubeConfig.APIVersion = kubeConfig.APIVersion
	cleanKubeConfig.Kind = kubeConfig.Kind
	cleanKubeConfig.Preferences = kubeConfig.Preferences
	cleanKubeConfig.Extensions = kubeConfig.Extensions
	cleanKubeConfig.CurrentContext = kubeConfig.CurrentContext

	return cleanKubeConfig, nil
}

func handleFailedHostDeletion(err error) {
	if forceFlag {
		err := os.RemoveAll(state.InstanceDirs.Machines)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error deleting '%s': %v", state.InstanceDirs.Machines, err))
		}
	} else {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error deleting the Minishift VM: %v", err))
	}
}

func removeInstanceAndKubeConfig() {
	exists := filehelper.Exists(minishiftConfig.InstanceStateConfig.FilePath)
	if exists {
		if err := minishiftConfig.InstanceStateConfig.Delete(); err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintln(fmt.Sprintf("Error deleting '%s': ", minishiftConfig.InstanceStateConfig.FilePath), err))
		}
	}
	exists = filehelper.Exists(constants.KubeConfigPath)
	if exists {
		if err := os.Remove(constants.KubeConfigPath); err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error deleting '%s'", constants.KubeConfigPath))
		}
	}
}

// deleteExistingDirectory delete the directory which minishift create in case of generic driver.
// As of now even after cluster down there are some mount point left which cause issue to delete entire
// directory tree so as of now ignoring the error [PK]
func deleteExistingDirectory(hostVm *host.Host) error {
	sshCommander := provision.GenericSSHCommander{Driver: hostVm.Driver}
	cmd := fmt.Sprintf("sudo rm -fr /var/lib/minishift/*")
	sshCommander.SSHCommand(cmd)
	return nil
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Forces the deletion of the VM specific files in MINISHIFT_HOME.")
	deleteCmd.Flags().BoolVar(&clearCacheFlag, "clear-cache", false, "Deletes all cached content. This affects all profiles.")
	RootCmd.AddCommand(deleteCmd)
}
