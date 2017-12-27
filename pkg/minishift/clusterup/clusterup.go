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

package clusterup

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/golang/glog"

	"os"
	"strings"
	"time"

	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minikube/kubeconfig"
	"github.com/minishift/minishift/pkg/minishift/addon/command"
	"github.com/minishift/minishift/pkg/minishift/addon/manager"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/minishift/oc"
	openshiftVersion "github.com/minishift/minishift/pkg/minishift/openshift/version"

	"regexp"

	"github.com/minishift/minishift/pkg/util"
	minishiftStrings "github.com/minishift/minishift/pkg/util/strings"
)

const (
	ipKey            = "ip"
	routingSuffixKey = "routing-suffix"
	envPrefix        = "env."
)

type ClusterUpConfig struct {
	OpenShiftVersion string
	MachineName      string
	Ip               string
	Port             int
	RoutingSuffix    string
	HostPvDir        string
	User             string
	Project          string
	KubeConfigPath   string
	OcPath           string
	AddonEnv         []string
	PublicHostname   string
}

// ClusterUp downloads and installs the oc binary in order to run 'cluster up'
func ClusterUp(config *ClusterUpConfig, clusterUpParams map[string]string, runner util.Runner) error {
	cmdArgs := []string{"cluster", "up", "--use-existing-config"}

	fmt.Println("-- Checking 'oc' support for startup flags ... ")

	// Deal with extra flags (remove from cluster up params)
	var extraFlags string
	if val, ok := clusterUpParams[configCmd.ExtraClusterUpFlags.Name]; ok {
		extraFlags = val
		delete(clusterUpParams, configCmd.ExtraClusterUpFlags.Name)
	}

	// Check if clusterUp flags are supported
	for key, value := range clusterUpParams {
		fmt.Printf("   %s ... ", key)
		if !oc.SupportFlag(key, config.OcPath, runner) {
			fmt.Println("FAIL")
			return errors.New(fmt.Sprintf("Flag '%s' is not supported for oc version %s. Use 'openshift-version' flag to select a different version of OpenShift.", key, config.OpenShiftVersion))
		}
		fmt.Println("OK")

		cmdArgs = append(cmdArgs, "--"+key)
		cmdArgs = append(cmdArgs, value)
	}

	if minishiftConfig.EnableExperimental {
		// Deal with extra flags (add to command arguments)
		if len(extraFlags) > 0 {
			fmt.Println("-- Extra 'oc' cluster up flags (experimental) ... ")
			fmt.Printf("   '%s'\n", extraFlags)
			extraFlagFields := strings.Fields(extraFlags)
			for _, extraFlag := range extraFlagFields {
				cmdArgs = append(cmdArgs, extraFlag)
			}
		}
	}

	if glog.V(2) {
		fmt.Printf("-- Running 'oc' with: '%s'\n", strings.Join(cmdArgs, " "))
	}
	exitCode := runner.Run(os.Stdout, os.Stderr, config.OcPath, cmdArgs...)
	if exitCode != 0 {
		return errors.New("Error starting the cluster.")
	}
	return nil
}

// PostClusterUp runs the Minishift specific provisioning after 'cluster up' has run
func PostClusterUp(clusterUpConfig *ClusterUpConfig, sshCommander provision.SSHCommander, addOnManager *manager.AddOnManager) error {
	// With oc 3.6 client, clusterName entry to kubeconfig file become 127.0.0.1:<port> and
	// we started to use oc baseline >= 3.6 which can provision now older release version.
	// In case there is any change happen to newer version of oc client binary then this need modification.
	userName := fmt.Sprintf("system:admin/127-0-0-1:%d", clusterUpConfig.Port)

	err := kubeconfig.CacheSystemAdminEntries(clusterUpConfig.KubeConfigPath, getConfigClusterName(clusterUpConfig.PublicHostname, clusterUpConfig.Ip, clusterUpConfig.Port), userName)
	if err != nil {
		return err
	}

	ocRunner, err := oc.NewOcRunner(clusterUpConfig.OcPath, clusterUpConfig.KubeConfigPath)
	if err != nil {
		return err
	}

	err = ocRunner.AddSudoerRoleForUser(clusterUpConfig.User)
	if err != nil {
		return err
	}

	err = ocRunner.AddCliContext(clusterUpConfig.MachineName, clusterUpConfig.Ip, clusterUpConfig.User, clusterUpConfig.Project)
	if err != nil {
		return err
	}

	err = configurePersistentVolumes(clusterUpConfig.HostPvDir, addOnManager, sshCommander, ocRunner)
	if err != nil {
		return err
	}

	err = applyAddOns(addOnManager, clusterUpConfig.Ip, clusterUpConfig.RoutingSuffix, clusterUpConfig.AddonEnv, ocRunner, sshCommander)
	if err != nil {
		return err
	}

	return nil
}

// EnsureHostDirectoriesExist ensures that the specified directories exist on the VM and creates them if not.
func EnsureHostDirectoriesExist(host *host.Host, dirs []string) error {
	cmd := fmt.Sprintf("sudo mkdir -p %s", strings.Join(dirs, " "))
	_, err := host.RunSSHCommand(cmd)
	if err != nil {
		return err
	}
	return nil
}

// DetermineOcVersion returns the oc version to use.
// If the requested version is < v3.7.0 we will use oc binary v3.6.0 to provision the requested OpenShift version.
// If the requested OpenShift version is >= v3.7.0, we align the oc version with the requested OpenShift version.
// Check Minishift github issue #1417 for details.
func DetermineOcVersion(requestedVersion string) string {
	valid, _ := openshiftVersion.IsGreaterOrEqualToBaseVersion(requestedVersion, constants.BackwardIncompatibleOcVersion)
	if !valid {
		requestedVersion = constants.MinimumOcBinaryVersion
	}

	return requestedVersion
}

// GetExecutionContext creates an ExecutionContext used for variable interpolation during add-on application.
// The context contains variables to interpolate during add-on execution, as well as the means to communicate with the VM (SSHCommander) and OpenShift (OcRunner).
func GetExecutionContext(ip string, routingSuffix string, addOnEnv []string, ocRunner *oc.OcRunner, sshCommander provision.SSHCommander) (*command.ExecutionContext, error) {
	context, err := command.NewExecutionContext(ocRunner, sshCommander)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to initialise execution context: %s", err.Error()))
	}

	context.AddToContext(ipKey, ip)
	context.AddToContext(routingSuffixKey, routingSuffix)

	for _, env := range addOnEnv {
		match, _ := regexp.Match(".*=.*", []byte(env))
		if !match {
			return nil, errors.New(fmt.Sprintf("Add-on interpolation variables need to be specified in the format <key>=<value>.'%s' is not.", env))
		}

		envTokens, err := minishiftStrings.SplitAndTrim(env, "=")
		if err != nil {
			return nil, err
		}

		key, value := envTokens[0], envTokens[1]
		if strings.HasPrefix(value, envPrefix) {
			if os.Getenv(strings.TrimPrefix(value, envPrefix)) != "" {
				value = os.Getenv(strings.TrimPrefix(value, envPrefix))
			} else {
				continue
			}
		}

		context.AddToContext(key, value)
	}

	return context, nil
}

// TODO - persistent volume creation should really be fixed upstream, aka 'cluster up'. See https://github.com/openshift/origin/issues/14076 (HF)
// configurePersistentVolumes makes sure that the default persistent volumes created by 'cluster up' have the right permissions - see https://github.com/minishift/minishift/issues/856
func configurePersistentVolumes(hostPvDir string, addOnManager *manager.AddOnManager, sshCommander provision.SSHCommander, ocRunner *oc.OcRunner) error {
	// don't apply this if anyuid is not enabled
	anyuid := addOnManager.Get("anyuid")
	if anyuid == nil || !anyuid.IsEnabled() {
		return nil
	}

	fmt.Print("-- Waiting for persistent volumes to be created ... ")

	var out, err *bytes.Buffer

	// poll the status of the persistent-volume-setup job to determine when the persitent volume creates is completed
	timeout := time.NewTimer(5 * time.Minute)
outerPollActive:
	for {
		select {
		case <-timeout.C:
			return errors.New("Timed out to poll active state of persistent-volume-setup job")
		default:
			out = new(bytes.Buffer)
			err = new(bytes.Buffer)
			exitStatus := ocRunner.Run("get job persistent-volume-setup -n default -o 'jsonpath={ .status.active }'", out, err)
			if exitStatus != 0 || len(err.String()) > 0 {
				return errors.New("Unable to monitor persistent volume creation")
			}

			if out.String() != "1" {
				break outerPollActive
			}

			time.Sleep(1 * time.Second)
		}
	}

	// poll the success status of persistent-volume-setup job.
outerPollSuccess:
	for {
		select {
		case <-timeout.C:
			return errors.New("Timed out to poll success state of persistent-volume-setup job")
		default:
			out = new(bytes.Buffer)
			err = new(bytes.Buffer)
			exitStatus := ocRunner.Run("get job persistent-volume-setup -n default -o 'jsonpath={ .status.succeeded }'", out, err)

			if exitStatus != 0 || len(err.String()) > 0 {
				return errors.New("Persistent volume creation failed")
			}

			if out.String() == "1" {
				break outerPollSuccess
			}

			time.Sleep(1 * time.Second)
		}
	}

	cmd := fmt.Sprintf("sudo chmod -R 777 %s/pv*", hostPvDir)
	sshCommander.SSHCommand(cmd)

	// if we have SELinux enabled we need to sort things out there as well
	// 'cluster up' does this as well, but we do it here as well to have all required actions collected in one
	// place, instead of relying on some implicit knowledge on what 'cluster up does (HF)
	cmd = fmt.Sprintf("sudo which chcon; if [ $? -eq 0 ]; then chcon -R -t svirt_sandbox_file_t %s/pv*; fi", hostPvDir)
	sshCommander.SSHCommand(cmd)

	cmd = fmt.Sprintf("sudo which restorecon; if [ $? -eq 0 ]; then restorecon -R %s/pv*; fi", hostPvDir)
	sshCommander.SSHCommand(cmd)

	fmt.Println("OK")
	fmt.Println()

	return nil
}

func applyAddOns(addOnManager *manager.AddOnManager, ip string, routingSuffix string, addonEnv []string, ocRunner *oc.OcRunner, sshCommander provision.SSHCommander) error {
	context, err := GetExecutionContext(ip, routingSuffix, addonEnv, ocRunner, sshCommander)
	if err != nil {
		return err
	}

	err = addOnManager.Apply(context)
	if err != nil {
		return err
	}

	return nil
}

func getConfigClusterName(hostname string, ip string, port int) string {
	if hostname != "" {
		return "local-cluster"
	}
	return fmt.Sprintf("%s:%d", strings.Replace(ip, ".", "-", -1), port)
}
