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
	"path/filepath"
	"strings"

	"os"

	"github.com/asaskevich/govalidator"
	units "github.com/docker/go-units"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/provision"
	"github.com/golang/glog"
	"github.com/minishift/minishift/cmd/minishift/cmd/addon"
	startFlags "github.com/minishift/minishift/cmd/minishift/cmd/config"
	cmdutil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/cache"
	"github.com/minishift/minishift/pkg/minishift/clusterup"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/minishift/minishift/pkg/minishift/docker/image"
	"github.com/minishift/minishift/pkg/minishift/hostfolder"
	"github.com/minishift/minishift/pkg/minishift/openshift"
	"github.com/minishift/minishift/pkg/minishift/provisioner"
	"github.com/minishift/minishift/pkg/util"
	inputUtils "github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/minishift/minishift/pkg/version"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	commandName = "start"

	defaultProject = "myproject"
	defaultUser    = "developer"

	defaultInsecureRegistry = "172.30.0.0/16"

	unsupportedIsoUrlFormat = "Unsupported value for iso-url. It can be an URL, file URI or one of the following short names: [b2d centos]."
)

var (
	dockerEnv        []string
	dockerEngineOpt  []string
	insecureRegistry []string
	registryMirror   []string
	openShiftEnv     []string
	shellProxyEnv    string
)

// startCmd represents the start command
var startCmd *cobra.Command

// Set default value for host data and config dir
var (
	hostConfigDirectory  = "/var/lib/minishift/openshift.local.config"
	hostDataDirectory    = "/var/lib/minishift/hostdata"
	hostVolumesDirectory = "/var/lib/minishift/openshift.local.volumes"
	hostPvDirectory      = "/var/lib/minishift/openshift.local.pv"
)

// clusterUpFlagSet contains the command line switches which needs to be passed on to 'cluster up'
var clusterUpFlagSet *flag.FlagSet

// minishiftToClusterUp is a mapping between flag names used in minishift CLI and flag name as passed to 'cluster up'
var minishiftToClusterUp = map[string]string{
	"openshift-env":     "env",
	"openshift-version": "version",
}

// init configures the command line options of this command
func init() {
	// need to initialize startCmd in init to avoid initialization loop (runStart calls validateOpenshiftVersion
	// which in turn makes use of startCmd
	startCmd = &cobra.Command{
		Use:   commandName,
		Short: "Starts a local OpenShift cluster.",
		Long: `Starts a local single-node OpenShift cluster.

All flags of this command can also be configured by setting corresponding environment variables or persistent configuration options.
For the former prefix the flag with MINISHIFT_, uppercase characters and replace '-' with '_', for example MINISHIFT_VM_DRIVER.
For the latter see 'minishift config -h'.`,
		Run: runStart,
	}

	clusterUpFlagSet = initClusterUpFlags()
	startCmd.Flags().AddFlagSet(clusterUpFlagSet)
	startCmd.Flags().AddFlagSet(initStartFlags())
	startCmd.Flags().AddFlagSet(initSubscriptionManagerFlags())

	viper.BindPFlags(startCmd.Flags())
	RootCmd.AddCommand(startCmd)

	provision.SetDetector(&provisioner.MinishiftProvisionerDetector{Delegate: provision.StandardDetector{}})
}

// runStart handles all command line arguments, launches the VM and provisions OpenShift
func runStart(cmd *cobra.Command, args []string) {
	libMachineClient := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer libMachineClient.Close()

	ensureNotRunning(libMachineClient, constants.MachineName)
	validateOpenshiftVersion()
	setSubscriptionManagerParameters()

	proxyConfig := handleProxies()

	// to determine whether we need to run post cluster up actions,
	// we need to determine whether this is a restart prior to potentially creating a new VM
	isRestart := cmdutil.VMExists(libMachineClient, constants.MachineName)

	hostVm, ip := startHost(libMachineClient)
	registerHost(libMachineClient)

	if proxyConfig.IsEnabled() {
		// once we know the IP, we need to make sure it is not proxied in a proxy environment
		proxyConfig.AddNoProxy(ip)
		proxyConfig.ApplyToEnvironment()
	}

	applyDockerEnvToProcessEnv(libMachineClient)

	err := clusterup.EnsureHostDirectoriesExist(hostVm, getRequiredHostDirectories())
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating required host directories: %v", err))
	}

	autoMountHostFolders(hostVm.Driver)

	requestedOpenShiftVersion := viper.GetString(startFlags.OpenshiftVersion.Name)
	if !isRestart {
		importContainerImages(hostVm, requestedOpenShiftVersion)
	}

	ocPath := cacheOc(clusterup.DetermineOcVersion(requestedOpenShiftVersion))
	clusterUpConfig := &clusterup.ClusterUpConfig{
		OpenShiftVersion: requestedOpenShiftVersion,
		MachineName:      constants.MachineName,
		Ip:               ip,
		Port:             constants.APIServerPort,
		RoutingSuffix:    getDefaultRoutingPrefix(ip),
		HostPvDir:        viper.GetString(startFlags.HostPvDir.Name),
		User:             defaultUser,
		Project:          defaultProject,
		KubeConfigPath:   constants.KubeConfigPath,
		OcPath:           ocPath,
		AddonEnv:         viper.GetStringSlice(cmdutil.AddOnEnv),
	}

	clusterUpParams := determineClusterUpParameters(clusterUpConfig)
	err = clusterup.ClusterUp(clusterUpConfig, clusterUpParams, &util.RealRunner{})

	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error during 'cluster up' execution: %v", err))
	}

	if !IsOpenShiftRunning(hostVm.Driver) {
		atexit.ExitWithMessage(1, "OpenShift provisioning failed. origin container failed to start.")
	}

	if !isRestart {
		postClusterUp(hostVm, clusterUpConfig)
		exportContainerImages(hostVm, requestedOpenShiftVersion)
	}
}

// postClusterUp performs configuration action which only need to be run after an initial provision of OpenShift.
// On subsequent VM restarts these actions can be skipped.
func postClusterUp(hostVm *host.Host, clusterUpConfig *clusterup.ClusterUpConfig) {
	sshCommander := provision.GenericSSHCommander{Driver: hostVm.Driver}
	err := clusterup.PostClusterUp(clusterUpConfig, sshCommander, addon.GetAddOnManager())
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error during post cluster up configuration: %v", err))
	}
}

// getRequiredHostDirectories returns a list of directories we need to ensure exist on the VM.
func getRequiredHostDirectories() []string {
	return []string{
		viper.GetString(startFlags.HostConfigDir.Name),
		viper.GetString(startFlags.HostDataDir.Name),
		viper.GetString(startFlags.HostVolumeDir.Name),
		viper.GetString(startFlags.HostPvDir.Name),
	}
}

func handleProxies() *util.ProxyConfig {
	proxyConfig, err := util.NewProxyConfig(viper.GetString(cmdutil.HttpProxy), viper.GetString(cmdutil.HttpsProxy), viper.GetString(startFlags.NoProxyList.Name))

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
			viper.Set(cmdutil.HttpProxy, proxyConfig.HttpProxy())
		}

		if proxyConfig.HttpsProxy() != "" {
			viper.Set(cmdutil.HttpsProxy, proxyConfig.HttpsProxy())
		}
		viper.Set(startFlags.NoProxyList.Name, proxyConfig.NoProxy())
	}

	return proxyConfig
}

// getSlice return slice for provided key if value type is []interface{} otherwise nil
func getSlice(key string) []string {
	if viper.IsSet(key) {
		value := viper.Get(key)
		if _, ok := value.([]interface{}); ok {
			return viper.GetStringSlice(key)
		}
	}
	return nil
}

func determineInsecureRegistry(key string) []string {
	s := getSlice(key)
	if s != nil {
		for _, v := range s {
			if v == defaultInsecureRegistry {
				return s
			}
		}
	}
	return append(s, defaultInsecureRegistry)
}

func startHost(libMachineClient *libmachine.Client) (*host.Host, string) {
	machineConfig := &cluster.MachineConfig{
		MinikubeISO:      determineIsoUrl(viper.GetString(startFlags.ISOUrl.Name)),
		Memory:           viper.GetInt(startFlags.Memory.Name),
		CPUs:             viper.GetInt(startFlags.CPUs.Name),
		DiskSize:         calculateDiskSizeInMB(viper.GetString(startFlags.DiskSize.Name)),
		VMDriver:         viper.GetString(startFlags.VmDriver.Name),
		DockerEnv:        append(dockerEnv, getSlice(startFlags.DockerEnv.Name)...),
		DockerEngineOpt:  append(dockerEngineOpt, getSlice(startFlags.DockerEngineOpt.Name)...),
		InsecureRegistry: append(insecureRegistry, determineInsecureRegistry(startFlags.InsecureRegistry.Name)...),
		RegistryMirror:   append(registryMirror, getSlice(startFlags.RegistryMirror.Name)...),
		HostOnlyCIDR:     viper.GetString(startFlags.HostOnlyCIDR.Name),
		ShellProxyEnv:    shellProxyEnv,
	}

	fmt.Printf("Starting local OpenShift cluster using '%s' hypervisor...\n", machineConfig.VMDriver)
	var hostVm *host.Host
	start := func() (err error) {
		hostVm, err = cluster.StartHost(libMachineClient, *machineConfig)
		if err != nil {
			glog.Errorf("Error starting the VM: %s. Retrying.\n", err)
		}
		return err
	}
	err := util.Retry(3, start)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error starting the VM: %v", err))
	}

	ip, err := hostVm.Driver.GetIP()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error determining host ip: %v ", err))
	}

	return hostVm, ip
}

func autoMountHostFolders(driver drivers.Driver) {
	if hostfolder.IsAutoMount() && hostfolder.IsHostfoldersDefined() {
		hostfolder.MountHostfolders(driver)
	}
}

func importContainerImages(hostVm *host.Host, openShiftVersion string) {
	if !viper.GetBool(startFlags.ImageCaching.Name) {
		return
	}

	handler := getImageHandler(hostVm)
	config := &image.ImageCacheConfig{
		HostCacheDir: constants.MakeMiniPath("cache", "images"),
		CachedImages: image.GetOpenShiftImageNames(openShiftVersion),
	}
	handler.ImportImages(config)
}

func getImageHandler(hostVm *host.Host) image.ImageHandler {
	handler, err := image.NewDockerImageHandler(hostVm.Driver)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Unable to create image handler: %v", err))
	}

	return handler
}

// exportContainerImages exports the OpenShift images in a background process (by calling 'minishift image export')
func exportContainerImages(hostVm *host.Host, version string) {
	if !viper.GetBool(startFlags.ImageCaching.Name) {
		return
	}

	handler := getImageHandler(hostVm)
	config := &image.ImageCacheConfig{
		HostCacheDir: constants.MakeMiniPath("cache", "images"),
		CachedImages: image.GetOpenShiftImageNames(version),
	}

	if handler.AreImagesCached(config) {
		return
	}

	exportCmd, err := image.CreateExportCommand(version)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating export command: %v", err))
	}

	err = exportCmd.Start()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error during export: %v", err))
	}
	fmt.Println(fmt.Sprintf("-- Exporting of OpenShift images is occuring in background process with pid %d.", exportCmd.Process.Pid))
}

// calculateDiskSizeInMB converts a human specified disk size like "1000MB" or "1GB" and converts it into Megabits
func calculateDiskSizeInMB(humanReadableDiskSize string) int {
	diskSize, err := units.FromHumanSize(humanReadableDiskSize)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Disk size is not valid: %v", err))
	}
	return int(diskSize / units.MB)
}

func determineIsoUrl(iso string) string {
	iso = strings.ToLower(iso)
	isoNotSpecified := ""

	switch iso {
	case startFlags.B2dIsoAlias, isoNotSpecified:
		iso = constants.DefaultB2dIsoUrl
	case startFlags.CentOsIsoAlias:
		iso = constants.DefaultCentOsIsoUrl
	default:
		if !(govalidator.IsURL(iso) || strings.HasPrefix(iso, "file:")) {
			atexit.ExitWithMessage(1, unsupportedIsoUrlFormat)
		}
	}

	return iso
}

// initStartFlags creates the CLI flags which needs to be passed on to 'libmachine'
func initStartFlags() *flag.FlagSet {
	startFlagSet := flag.NewFlagSet(commandName, flag.ContinueOnError)

	startFlagSet.String(startFlags.ISOUrl.Name, startFlags.B2dIsoAlias, "Location of the minishift ISO. Can be an URL, file URI or one of the following short names: [b2d centos].")
	startFlagSet.String(startFlags.VmDriver.Name, constants.DefaultVMDriver, fmt.Sprintf("The driver to use for the Minishift VM. Possible values: %v", constants.SupportedVMDrivers))
	startFlagSet.Int(startFlags.Memory.Name, constants.DefaultMemory, "Amount of RAM to allocate to the Minishift VM.")
	startFlagSet.Int(startFlags.CPUs.Name, constants.DefaultCPUS, "Number of CPU cores to allocate to the Minishift VM.")
	startFlagSet.String(startFlags.DiskSize.Name, constants.DefaultDiskSize, "Disk size to allocate to the Minishift VM. Use the format <size><unit>, where unit = b, k, m or g.")
	startFlagSet.String(startFlags.HostOnlyCIDR.Name, "192.168.99.1/24", "The CIDR to be used for the minishift VM. (Only supported with VirtualBox driver.)")
	startFlagSet.StringArrayVar(&dockerEnv, startFlags.DockerEnv.Name, nil, "Environment variables to pass to the Docker daemon. Use the format <key>=<value>.")
	startFlagSet.StringSliceVar(&dockerEngineOpt, startFlags.DockerEngineOpt.Name, nil, "Specify arbitrary flags to pass to the Docker daemon in the form <flag>=<value>.")
	startFlagSet.StringSliceVar(&insecureRegistry, startFlags.InsecureRegistry.Name, []string{defaultInsecureRegistry}, "Non-secure Docker registries to pass to the Docker daemon.")
	startFlagSet.StringSliceVar(&registryMirror, startFlags.RegistryMirror.Name, nil, "Registry mirrors to pass to the Docker daemon.")
	startFlagSet.AddFlag(cmdutil.AddOnEnvFlag)

	return startFlagSet
}

// initClusterUpFlags creates the CLI flags which needs to be passed on to 'oc cluster up'
func initClusterUpFlags() *flag.FlagSet {
	clusterUpFlagSet := flag.NewFlagSet(commandName, flag.ContinueOnError)

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
	clusterUpFlagSet.AddFlag(cmdutil.HttpProxyFlag)
	clusterUpFlagSet.AddFlag(cmdutil.HttpsProxyFlag)

	return clusterUpFlagSet
}

// initSubscriptionManagerFlags create the CLI flags which are needed for VM registration
func initSubscriptionManagerFlags() *flag.FlagSet {
	subscriptionManagerFlagSet := flag.NewFlagSet(commandName, flag.ContinueOnError)

	subscriptionManagerFlagSet.String(startFlags.Username.Name, "", "Username for the virtual machine registration.")
	subscriptionManagerFlagSet.String(startFlags.Password.Name, "", "Password for the virtual machine registration.")
	subscriptionManagerFlagSet.Bool(startFlags.SkipRegistration.Name, false, "Skip the virtual machine registration.")

	return subscriptionManagerFlagSet
}

// determineClusterUpParameters returns a map of flag names and values for the cluster up call.
func determineClusterUpParameters(config *clusterup.ClusterUpConfig) map[string]string {
	clusterUpParams := make(map[string]string)

	// Set default value for host config, data and volumes
	viper.Set(startFlags.HostConfigDir.Name, viper.GetString(startFlags.HostConfigDir.Name))
	viper.Set(startFlags.HostDataDir.Name, viper.GetString(startFlags.HostDataDir.Name))
	viper.Set(startFlags.HostVolumeDir.Name, viper.GetString(startFlags.HostVolumeDir.Name))
	viper.Set(startFlags.HostPvDir.Name, viper.GetString(startFlags.HostPvDir.Name))
	viper.Set(startFlags.RoutingSuffix.Name, config.RoutingSuffix)

	clusterUpFlagSet.VisitAll(func(flag *flag.Flag) {
		if viper.IsSet(flag.Name) {
			value := viper.GetString(flag.Name)
			key := flag.Name
			_, exists := minishiftToClusterUp[key]
			if exists {
				key = minishiftToClusterUp[key]
			}
			clusterUpParams[key] = value
		}
	})

	return clusterUpParams
}

// cacheOc ensures that the oc binary matching the requested OpenShift version is cached on the host
func cacheOc(openShiftVersion string) string {
	ocBinary := cache.Oc{
		OpenShiftVersion:  openShiftVersion,
		MinishiftCacheDir: filepath.Join(constants.Minipath, "cache"),
	}
	if err := ocBinary.EnsureIsCached(); err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error starting the cluster: %v", err))
	}

	// Update MACHINE_NAME.json for oc path
	minishiftConfig.InstanceConfig.OcPath = filepath.Join(ocBinary.GetCacheFilepath(), constants.OC_BINARY_NAME)
	if err := minishiftConfig.InstanceConfig.Write(); err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error updating oc path in config of VM: %v", err))
	}

	return minishiftConfig.InstanceConfig.OcPath
}

func getDefaultRoutingPrefix(ip string) string {
	// prefer nip.io over xip.io. See GitHub issue #501
	if viper.IsSet(startFlags.RoutingSuffix.Name) {
		return viper.GetString(startFlags.RoutingSuffix.Name)
	} else {
		return ip + ".nip.io"
	}
}

func ensureNotRunning(client *libmachine.Client, machineName string) {
	if !cmdutil.VMExists(client, machineName) {
		return
	}

	hostVm, err := client.Load(constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	if cmdutil.IsHostRunning(hostVm.Driver) {
		atexit.ExitWithMessage(0, fmt.Sprintf("The '%s' VM is already running.", machineName))
	}
}

func validateOpenshiftVersion() {
	requestedVersion := viper.GetString(startFlags.OpenshiftVersion.Name)

	valid, err := clusterup.ValidateOpenshiftMinVersion(requestedVersion, constants.MinOpenshiftSupportedVersion)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	if !valid {
		fmt.Printf("Minishift does not support OpenShift version %s. "+
			"You need to use a version >= %s\n", viper.GetString(startFlags.OpenshiftVersion.Name),
			constants.MinOpenshiftSupportedVersion)
		atexit.Exit(1)
	}

	// Make sure the version actually has a 'v' prefix. See https://github.com/minishift/minishift/issues/410
	if !strings.HasPrefix(requestedVersion, constants.VersionPrefix) {
		requestedVersion = constants.VersionPrefix + requestedVersion
		// this will make sure the right version is set in case the version comes from config file
		viper.Set(startFlags.OpenshiftVersion.Name, requestedVersion)

		// if the version was specified via the CLI we need to update the flag value
		startCmd.Flags().Lookup(startFlags.OpenshiftVersion.Name).Value.Set(requestedVersion)
	}
}

func setSubscriptionManagerParameters() {
	cluster.RegistrationParameters.Username = viper.GetString(startFlags.Username.Name)
	cluster.RegistrationParameters.Password = viper.GetString(startFlags.Password.Name)
	cluster.RegistrationParameters.GetUsernameInteractive = getUsernameInteractive
	cluster.RegistrationParameters.GetPasswordInteractive = getPasswordInteractive
}

func getUsernameInteractive(message string) string {
	return inputUtils.ReadInputFromStdin(message)
}

func getPasswordInteractive(message string) string {
	return inputUtils.ReadPasswordFromStdin(message)
}

func registerHost(libMachineClient *libmachine.Client) {
	if viper.GetBool(startFlags.SkipRegistration.Name) {
		log.Debug("Skipping registration due to enabled --skip-registration flag")
		return
	}
	supportRegistration, err := cluster.Register(libMachineClient)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error to register VM: %v", err))
	}
	if supportRegistration {
		minishiftConfig.InstanceConfig.IsRegistered = true
		minishiftConfig.InstanceConfig.Write()
		return
	}
}

func applyDockerEnvToProcessEnv(libMachineClient *libmachine.Client) {
	// Making sure the required Docker environment variables are set to make 'cluster up' work
	envMap, err := cluster.GetHostDockerEnv(libMachineClient)
	for k, v := range envMap {
		os.Setenv(k, v)
	}
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error determining Docker settings: %v", err))
	}
}

func IsOpenShiftRunning(driver drivers.Driver) bool {
	sshCommander := provision.GenericSSHCommander{Driver: driver}
	dockerCommander := docker.NewVmDockerCommander(sshCommander)

	return openshift.IsRunning(dockerCommander)
}
