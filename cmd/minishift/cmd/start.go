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

	"bytes"
	"time"

	units "github.com/docker/go-units"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/golang/glog"
	"github.com/minishift/minishift/cmd/minishift/cmd/addon"
	startFlags "github.com/minishift/minishift/cmd/minishift/cmd/config"
	cmdutil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minikube/kubeconfig"
	"github.com/minishift/minishift/pkg/minishift/addon/manager"
	"github.com/minishift/minishift/pkg/minishift/cache"
	"github.com/minishift/minishift/pkg/minishift/clusterup"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/minishift/hostfolder"
	"github.com/minishift/minishift/pkg/minishift/oc"
	"github.com/minishift/minishift/pkg/minishift/provisioner"
	minishiftUtil "github.com/minishift/minishift/pkg/minishift/util"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/minishift/minishift/pkg/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	commandName = "start"
)

var (
	dockerEnv        []string
	insecureRegistry []string
	registryMirror   []string
	openShiftEnv     []string
	shellProxyEnv    string
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   commandName,
	Short: "Starts a local OpenShift cluster.",
	Long:  `Starts a local single-node OpenShift cluster on the specified hypervisor.`,
	Run:   runStart,
}

// Set default value for host data and config dir
var (
	hostConfigDirectory  = "/var/lib/minishift/openshift.local.config"
	hostDataDirectory    = "/var/lib/minishift/hostdata"
	hostVolumesDirectory = "/var/lib/minishift/openshift.local.volumes"
	hostPvDirectory      = "/var/lib/minishift/openshift.local.pv"
)

// startFlagSet contains the minishift specific command line switches
var startFlagSet = flag.NewFlagSet(commandName, flag.ContinueOnError)

// clusterUpFlagSet contains the command line switches which needs to be passed on to 'cluster up'
var clusterUpFlagSet = flag.NewFlagSet(commandName, flag.ContinueOnError)

// subscription Manager FlagSet contains username and password details
var subscriptionManagerFlagSet = flag.NewFlagSet(commandName, flag.ContinueOnError)

// minishiftToClusterUp is a mapping between flag names used in minishift CLI and flag name as passed to 'cluster up'
var minishiftToClusterUp = map[string]string{
	"openshift-env":     "env",
	"openshift-version": "version",
}

// runner executes commands on the host
var runner util.Runner = &util.RealRunner{}

func SetRunner(newRunner util.Runner) {
	runner = newRunner
}

// runStart is executed as part of the start command
func runStart(cmd *cobra.Command, args []string) {
	libMachineClient := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer libMachineClient.Close()

	validateOpenshiftVersion()
	setSubscriptionManagerParameters()

	proxyConfig := handleProxies()

	machineConfig := cluster.MachineConfig{
		MinikubeISO:      viper.GetString(startFlags.ISOUrl.Name),
		Memory:           viper.GetInt(startFlags.Memory.Name),
		CPUs:             viper.GetInt(startFlags.CPUs.Name),
		DiskSize:         calculateDiskSizeInMB(viper.GetString(startFlags.DiskSize.Name)),
		VMDriver:         viper.GetString(startFlags.VmDriver.Name),
		DockerEnv:        dockerEnv,
		InsecureRegistry: insecureRegistry,
		RegistryMirror:   registryMirror,
		HostOnlyCIDR:     viper.GetString(startFlags.HostOnlyCIDR.Name),
		OpenShiftVersion: viper.GetString(startFlags.OpenshiftVersion.Name),
		ShellProxyEnv:    shellProxyEnv,
	}

	fmt.Printf("Starting local OpenShift cluster using '%s' hypervisor...\n", machineConfig.VMDriver)

	isRestart := cmdutil.VMExists(libMachineClient, constants.MachineName)

	var host *host.Host
	start := func() (err error) {
		host, err = cluster.StartHost(libMachineClient, machineConfig)
		if err != nil {
			glog.Errorf("Error starting the VM: %s. Retrying.\n", err)
		}
		return err
	}
	err := util.Retry(3, start)
	if err != nil {
		fmt.Println("Error starting the VM: ", err)
		atexit.Exit(1)
	}

	// Making sure the required Docker environment variables are set to make 'cluster up' work
	envMap, err := cluster.GetHostDockerEnv(libMachineClient)
	for k, v := range envMap {
		os.Setenv(k, v)
	}

	//Create the host directories if not present
	hostDirs := []string{viper.GetString(startFlags.HostConfigDir.Name),
		viper.GetString(startFlags.HostDataDir.Name), viper.GetString(startFlags.HostVolumeDir.Name),
		viper.GetString(startFlags.HostPvDir.Name)}
	err = clusterup.EnsureHostDirectoriesExist(libMachineClient, hostDirs)
	if err != nil {
		fmt.Println("Error creating required host directories: ", err)
		atexit.Exit(1)
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		fmt.Println("Error determining host ip: ", err)
		atexit.Exit(1)
	}

	if proxyConfig.IsEnabled() {
		// once we know the IP, we need to make sure it is not proxied in a proxy environment
		proxyConfig.AddNoProxy(ip)
		proxyConfig.ApplyToEnvironment()
	}

	automountHostfolders(host.Driver)

	clusterUp(&machineConfig, ip)

	if !isRestart {
		sshCommander := provision.GenericSSHCommander{Driver: host.Driver}
		postClusterUp(constants.MachineName, ip, constants.APIServerPort,
			viper.GetString(startFlags.RoutingSuffix.Name), minishiftConfig.InstanceConfig.OcPath,
			constants.KubeConfigPath, "developer", "myproject", sshCommander)
	}
}

func handleProxies() *util.ProxyConfig {
	proxyConfig, err := util.NewProxyConfig(viper.GetString(httpProxy), viper.GetString(httpsProxy),
		viper.GetString(startFlags.NoProxyList.Name))

	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	if proxyConfig.IsEnabled() {
		proxyConfig.ApplyToEnvironment()
		dockerEnv = append(dockerEnv, proxyConfig.ProxyConfig()...)
		shellProxyEnv += strings.Join(proxyConfig.ProxyConfig(), " ")

		// It could be that the proxy config is retrieved from the environment. To make sure that
		// proxy settings are properly passed to cluster up we need to explicitly set the values.
		if proxyConfig.HttpProxy() != "" {
			viper.Set(httpProxy, proxyConfig.HttpProxy())
		}

		if proxyConfig.HttpsProxy() != "" {
			viper.Set(httpsProxy, proxyConfig.HttpsProxy())
		}
		viper.Set(startFlags.NoProxyList.Name, proxyConfig.NoProxy())
	}

	return proxyConfig
}

// postClusterUp runs the Minishift specific provisioning after cluster up has run
func postClusterUp(machineName string, ip string, port int, routingSuffix string, ocPath string, kubeConfigPath string, user string, project string, sshCommander provision.SSHCommander) {
	if err := kubeconfig.CacheSystemAdminEntries(kubeConfigPath, getConfigClusterName(ip, port)); err != nil {
		fmt.Println("Error creating Minishift kubeconfig: ", err)
		atexit.Exit(1)
	}

	ocRunner, err := oc.NewOcRunner(ocPath, kubeConfigPath)
	if err != nil {
		fmt.Println("Error configuring OpenShift: ", err)
		atexit.Exit(1)
	}

	if err := ocRunner.AddSudoerRoleForUser(user); err != nil {
		glog.Error(fmt.Sprintf("Error giving %s sudoer privileges: ", user))
		atexit.Exit(1)
	}

	if err := ocRunner.AddCliContext(machineName, ip, user, project); err != nil {
		fmt.Println("Error adding OpenShift context: ", err)
		atexit.Exit(1)
	}

	// Run the default addons if MINISHIFT_HOME dir does not exist or empty
	if !homeDirExists || filehelper.IsDirEmpty(constants.Minipath) {
		fmt.Print("-- Installing default addons ... ")
		addon.UnpackAddons(constants.MakeMiniPath("addons"))
		fmt.Println("OK")
	}

	addOnManager := addon.GetAddOnManager()
	configurePersistentVolumes(addOnManager, sshCommander, ocRunner)
	applyAddOns(addOnManager, ip, routingSuffix, ocPath, kubeConfigPath, sshCommander)
}

func applyAddOns(addOnManager *manager.AddOnManager, ip string, routingSuffix string, ocPath string, kubeConfigPath string, sshCommander provision.SSHCommander) {
	err := addOnManager.Apply(addon.GetExecutionContext(ip, routingSuffix, ocPath, kubeConfigPath, sshCommander))
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprint("Error executing addon commands: ", err))
	}
}

// TODO - persistent volume creation should really be fixed upstream, aka 'cluster up'. See https://github.com/openshift/origin/issues/14076 (HF)
// configurePersistentVolumes makes sure that the default persistent volumes created by 'cluster up' have the right permissions - see https://github.com/minishift/minishift/issues/856
func configurePersistentVolumes(addOnManager *manager.AddOnManager, sshCommander provision.SSHCommander, ocRunner *oc.OcRunner) error {
	// don't apply this if anyuid is not enabled
	anyuid := addOnManager.Get("anyuid")
	if anyuid == nil || !anyuid.IsEnabled() {
		return nil
	}

	fmt.Print("-- Waiting for persistent volumes to be created ... ")

	hostPvDir := viper.GetString(startFlags.HostPvDir.Name)

	var out, err *bytes.Buffer

	// poll the status of the persistent-volume-setup job to determine when the persitent volume creates is completed
	for {
		out = new(bytes.Buffer)
		err = new(bytes.Buffer)
		exitStatus := ocRunner.Run("get job persistent-volume-setup -n default -o 'jsonpath={ .status.active }'", out, err)

		if exitStatus != 0 || len(err.String()) > 0 {
			return errors.New("Unable to monitor persistent volume creation")
		}

		if out.String() != "1" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	// verify the job succeeded
	out = new(bytes.Buffer)
	err = new(bytes.Buffer)
	exitStatus := ocRunner.Run("get job persistent-volume-setup -n default -o 'jsonpath={ .status.succeeded }'", out, err)

	if exitStatus != 0 || len(err.String()) > 0 || out.String() != "1" {
		return errors.New("Persistent volume creation failed")
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

func automountHostfolders(driver drivers.Driver) {
	if hostfolder.IsAutoMount() && hostfolder.IsHostfoldersDefined() {
		hostfolder.MountHostfolders(driver)
	}
}

// calculateDiskSizeInMB converts a human specified disk size like "1000MB" or "1GB" and converts it into Megabits
func calculateDiskSizeInMB(humanReadableDiskSize string) int {
	diskSize, err := units.FromHumanSize(humanReadableDiskSize)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Disk size is not valid: %s", err))
	}
	return int(diskSize / units.MB)
}

// init configures the command line options of this command
func init() {
	initStartFlags()
	initClusterUpFlags()
	initSubscriptionManagerFlags()

	startCmd.Flags().AddFlagSet(startFlagSet)
	startCmd.Flags().AddFlagSet(clusterUpFlagSet)
	startCmd.Flags().AddFlagSet(subscriptionManagerFlagSet)

	viper.BindPFlags(startCmd.Flags())
	RootCmd.AddCommand(startCmd)

	provision.SetDetector(&provisioner.MinishiftProvisionerDetector{Delegate: provision.StandardDetector{}})
}

// initStartFlags creates the CLI flags which needs to be passed on to 'libmachine'
func initStartFlags() {
	startFlagSet.String(startFlags.ISOUrl.Name, constants.DefaultIsoUrl, "Location of the minishift ISO.")
	startFlagSet.String(startFlags.VmDriver.Name, constants.DefaultVMDriver, fmt.Sprintf("The driver to use for the Minishift VM. Possible values: %v", constants.SupportedVMDrivers))
	startFlagSet.Int(startFlags.Memory.Name, constants.DefaultMemory, "Amount of RAM to allocate to the Minishift VM.")
	startFlagSet.Int(startFlags.CPUs.Name, constants.DefaultCPUS, "Number of CPU cores to allocate to the Minishift VM.")
	startFlagSet.String(startFlags.DiskSize.Name, constants.DefaultDiskSize, "Disk size to allocate to the Minishift VM. Use the format <size><unit>, where unit = b, k, m or g.")
	startFlagSet.String(startFlags.HostOnlyCIDR.Name, "192.168.99.1/24", "The CIDR to be used for the minishift VM. (Only supported with VirtualBox driver.)")
	startFlagSet.StringArrayVar(&dockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. Use the format <key>=<value>.")
	startFlagSet.StringSliceVar(&insecureRegistry, "insecure-registry", []string{"172.30.0.0/16"}, "Non-secure Docker registries to pass to the Docker daemon.")
	startFlagSet.StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon.")
}

// initClusterUpFlags creates the CLI flags which needs to be passed on to 'oc cluster up'
func initClusterUpFlags() {
	//clusterUpFlagSet.StringVar(&clusterUpConfig.Image, "image", "openshift/origin", "Specify the images to use for OpenShift")
	clusterUpFlagSet.Bool(startFlags.SkipRegistryCheck.Name, false, "Skip the Docker daemon registry check.")
	clusterUpFlagSet.String(startFlags.PublicHostname.Name, "", "Public hostname of the OpenShift cluster.")
	clusterUpFlagSet.String(startFlags.RoutingSuffix.Name, "", "Default suffix for the server routes.")
	clusterUpFlagSet.String(startFlags.HostConfigDir.Name, hostConfigDirectory, "Location of the OpenShift configuration on the Docker host.")
	clusterUpFlagSet.String(startFlags.HostVolumeDir.Name, hostVolumesDirectory, "Location of the OpenShift volumes on the Docker host.")
	clusterUpFlagSet.String(startFlags.HostDataDir.Name, hostDataDirectory, "Location of the OpenShift data on the Docker host. If not specified, etcd data will not be persisted on the host.")
	clusterUpFlagSet.String(startFlags.HostPvDir.Name, hostPvDirectory, "Directory on Docker host for OpenShift persistent volumes")
	clusterUpFlagSet.Int(startFlags.ServerLogLevel.Name, 0, "Log level for the OpenShift server.")
	clusterUpFlagSet.StringSliceVarP(&openShiftEnv, startFlags.OpenshiftEnv.Name, "e", []string{}, "Specify key-value pairs of environment variables to set on the OpenShift container.")
	clusterUpFlagSet.Bool(startFlags.Metrics.Name, false, "Install metrics (experimental)")
	clusterUpFlagSet.Bool(startFlags.Logging.Name, false, "Install logging (experimental)")
	clusterUpFlagSet.String(startFlags.OpenshiftVersion.Name, version.GetOpenShiftVersion(), fmt.Sprintf("The OpenShift version to run, eg. %s", version.GetOpenShiftVersion()))
	clusterUpFlagSet.String(startFlags.NoProxyList.Name, "", "List of hosts or subnets for which no proxy should be used.")
	clusterUpFlagSet.AddFlag(httpProxyFlag)
	clusterUpFlagSet.AddFlag(httpsProxyFlag)
}

// initProxyFlags create the CLI flags which needs to be passed for proxy
func initSubscriptionManagerFlags() {
	subscriptionManagerFlagSet.String(startFlags.Username.Name, "", "Username for the virtual machine registration.")
	subscriptionManagerFlagSet.String(startFlags.Password.Name, "", "Password for the virtual machine registration.")
}

// clusterUp downloads and installs the oc binary in order to run 'cluster up'
func clusterUp(config *cluster.MachineConfig, ip string) {
	if !minishiftUtil.ValidateOpenshiftMinVersion(viper.GetString(startFlags.OpenshiftVersion.Name), version.GetOpenShiftVersion()) {
		config.OpenShiftVersion = version.GetOpenShiftVersion()
	}
	oc := cache.Oc{
		OpenShiftVersion:  config.OpenShiftVersion,
		MinishiftCacheDir: filepath.Join(constants.Minipath, "cache"),
	}
	if err := oc.EnsureIsCached(); err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintln("Error starting the cluster: ", err))
	}

	// Update MACHINE_NAME.json for oc path
	minishiftConfig.InstanceConfig.OcPath = filepath.Join(oc.GetCacheFilepath(), constants.OC_BINARY_NAME)
	if err := minishiftConfig.InstanceConfig.Write(); err != nil {
		fmt.Println("Error updating oc path in config of VM: ", err)
		atexit.Exit(1)
	}

	cmdName := filepath.Join(oc.GetCacheFilepath(), constants.OC_BINARY_NAME)
	cmdArgs := []string{"cluster", "up", "--use-existing-config"}

	// Set default value for host config, data and volumes
	viper.Set(startFlags.HostConfigDir.Name, viper.GetString(startFlags.HostConfigDir.Name))
	viper.Set(startFlags.HostDataDir.Name, viper.GetString(startFlags.HostDataDir.Name))
	viper.Set(startFlags.HostVolumeDir.Name, viper.GetString(startFlags.HostVolumeDir.Name))
	viper.Set(startFlags.HostPvDir.Name, viper.GetString(startFlags.HostPvDir.Name))

	setDefaultRoutingPrefix(ip)

	clusterUpFlagSet.VisitAll(func(flag *flag.Flag) {
		if viper.IsSet(flag.Name) {
			value := viper.GetString(flag.Name)
			key := flag.Name
			_, exists := minishiftToClusterUp[key]
			if exists {
				key = minishiftToClusterUp[key]
			}
			if !ocSupportFlag(cmdName, key) {
				atexit.ExitWithMessage(1, fmt.Sprintf("Flag %s is not supported for oc version %s. Use 'openshift-version' flag to select a different version of OpenShift.", flag.Name, config.OpenShiftVersion))
			}
			cmdArgs = append(cmdArgs, "--"+key)
			cmdArgs = append(cmdArgs, value)
		}
	})

	exitCode := runner.Run(os.Stdout, os.Stderr, cmdName, cmdArgs...)
	if exitCode != 0 {
		atexit.ExitWithMessage(1, "Error starting the cluster.")
	}
}

func setDefaultRoutingPrefix(ip string) {
	// prefer nip.io over xip.io. See GitHub issue #501
	if !viper.IsSet(startFlags.RoutingSuffix.Name) {
		viper.Set(startFlags.RoutingSuffix.Name, ip+".nip.io")
	}
}

func validateOpenshiftVersion() {
	if viper.IsSet(startFlags.OpenshiftVersion.Name) {
		if !minishiftUtil.ValidateOpenshiftMinVersion(viper.GetString(startFlags.OpenshiftVersion.Name), constants.MinOpenshiftSuportedVersion) {
			fmt.Printf("Minishift does not support Openshift version %s ."+
				"You need to use a version >=%s\n", viper.GetString(startFlags.OpenshiftVersion.Name),
				constants.MinOpenshiftSuportedVersion)
			atexit.Exit(1)
		}
	}
}

func ocSupportFlag(cmdName string, flag string) bool {
	cmdArgs := []string{"cluster", "up", "-h"}
	cmdOut, err := runner.Output(cmdName, cmdArgs...)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Not able to get output of 'oc -h' Error: %s", err))
	}
	ocCommandOptions := minishiftUtil.ParseOcHelpCommand(cmdOut)
	if ocCommandOptions != nil {
		return minishiftUtil.FlagExist(ocCommandOptions, flag)
	}
	return false
}

func setSubscriptionManagerParameters() {
	cluster.RegistrationParameters.Username = viper.GetString(startFlags.Username.Name)
	cluster.RegistrationParameters.Password = viper.GetString(startFlags.Password.Name)
}

func getConfigClusterName(ip string, port int) string {
	return fmt.Sprintf("%s:%d", strings.Replace(ip, ".", "-", -1), port)
}
