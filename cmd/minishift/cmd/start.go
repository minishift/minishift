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
	"runtime"
	"strings"

	"os"

	"github.com/asaskevich/govalidator"
	units "github.com/docker/go-units"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/golang/glog"
	"github.com/minishift/minishift/cmd/minishift/cmd/addon"
	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	registrationUtil "github.com/minishift/minishift/cmd/minishift/cmd/registration"
	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftCluster "github.com/minishift/minishift/pkg/minishift/cluster"
	"github.com/minishift/minishift/pkg/minishift/clusterup"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/minishift/minishift/pkg/minishift/docker/image"
	"github.com/minishift/minishift/pkg/minishift/hostfolder"
	minishiftNetwork "github.com/minishift/minishift/pkg/minishift/network"
	"github.com/minishift/minishift/pkg/minishift/openshift"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/minishift/provisioner"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/minishift/minishift/pkg/util/progressdots"
	stringUtils "github.com/minishift/minishift/pkg/util/strings"
	"github.com/minishift/minishift/pkg/version"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	commandName             = "start"
	defaultInsecureRegistry = "172.30.0.0/16"
)

var (
	dockerEnv               []string
	openShiftEnv            []string
	shellProxyEnv           string
	unsupportedIsoUrlFormat = fmt.Sprintf("Unsupported value for iso-url. It can be a URL, file URI or one of the following aliases: [%s].",
		strings.Join(minishiftConstants.ValidIsoAliases, ","))

	// custom flags variable
	dockerEnvFlag = &flag.Flag{
		Name:      configCmd.DockerEnv.Name,
		Shorthand: "",
		Usage:     "Environment variables to pass to the Docker daemon. Use the format <key>=<value>.",
		Value:     cmdUtil.NewStringSliceValue([]string{}, &[]string{}),
	}

	insecureRegistryFlag = &flag.Flag{
		Name:      configCmd.InsecureRegistry.Name,
		Shorthand: "",
		Usage:     "Non-secure Docker registries to pass to the Docker daemon.",
		Value:     cmdUtil.NewStringSliceValue([]string{defaultInsecureRegistry}, &[]string{}),
	}

	dockerEngineOptFlag = &flag.Flag{
		Name:      configCmd.DockerEngineOpt.Name,
		Shorthand: "",
		Usage:     "Specify arbitrary flags to pass to the Docker daemon in the form <flag>=<value>.",
		Value:     cmdUtil.NewStringSliceValue([]string{}, &[]string{}),
	}

	registryMirrorFlag = &flag.Flag{
		Name:      configCmd.RegistryMirror.Name,
		Shorthand: "",
		Usage:     "Registry mirrors to pass to the Docker daemon.",
		Value:     cmdUtil.NewStringSliceValue([]string{}, &[]string{}),
	}

	nameServersFlag = &flag.Flag{
		Name:      configCmd.NameServers.Name,
		Shorthand: "",
		Usage:     "Specify nameserver to use for the instance.",
		Value:     cmdUtil.NewStringSliceValue([]string{}, &[]string{}),
	}

	startCmd *cobra.Command

	// Set default value for host data and config dir
	dataDirectory = map[string]string{
		"host-config-dir":  "/var/lib/minishift/openshift.local.config",
		"host-data-dir":    "/var/lib/minishift/hostdata",
		"host-volumes-dir": "/var/lib/minishift/openshift.local.volumes",
		"host-pv-dir":      "/var/lib/minishift/openshift.local.pv",
	}

	// clusterUpFlagSet contains the command line switches which needs to be passed on to 'cluster up'
	clusterUpFlagSet *flag.FlagSet

	// ocPath
	ocPath = ""
)

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
	fmt.Println(fmt.Sprintf("-- Starting profile '%s'", constants.ProfileName))

	libMachineClient := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer libMachineClient.Close()

	ensureNotRunning(libMachineClient, constants.MachineName)
	addVersionPrefixToOpenshiftVersion()

	// to determine whether we need to run post cluster up actions,
	// we need to determine whether this is a restart prior to potentially creating a new VM
	isRestart := cmdUtil.VMExists(libMachineClient, constants.MachineName)

	proxyConfig := handleProxies()

	// preflight check (before start)
	preflightChecksBeforeStartingHost()

	// Cache OC binary before starting the VM and perform oc command option check
	requestedOpenShiftVersion := viper.GetString(configCmd.OpenshiftVersion.Name)
	ocPath = cmdUtil.CacheOc(requestedOpenShiftVersion)
	preflightChecksForArtifacts()

	setSubscriptionManagerParameters()

	fmt.Print("-- Starting local OpenShift cluster")

	hostVm := startHost(libMachineClient)
	registrationUtil.RegisterHost(libMachineClient)

	// Forcibly set nameservers when configured
	minishiftNetwork.AddNameserversToInstance(hostVm.Driver, getSlice(configCmd.NameServers.Name))

	// preflight checks (after start)
	preflightChecksAfterStartingHost(hostVm.Driver)

	// Adding active profile information to all instance config
	addActiveProfileInformation()

	ip, _ := hostVm.Driver.GetIP()

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

	if !isNoProvision() {
		if !isRestart {
			importContainerImages(hostVm.Driver, libMachineClient, requestedOpenShiftVersion)
		}

		sshCommander := provision.GenericSSHCommander{Driver: hostVm.Driver}
		dockerCommander := docker.NewVmDockerCommander(sshCommander)

		clusterUpConfig := &clusterup.ClusterUpConfig{
			OpenShiftVersion:     requestedOpenShiftVersion,
			MachineName:          constants.MachineName,
			Ip:                   ip,
			Port:                 constants.APIServerPort,
			RoutingSuffix:        configCmd.GetDefaultRoutingSuffix(ip),
			User:                 minishiftConstants.DefaultUser,
			Project:              minishiftConstants.DefaultProject,
			KubeConfigPath:       constants.KubeConfigPath,
			OcPath:               ocPath,
			AddonEnv:             viper.GetStringSlice(cmdUtil.AddOnEnv),
			PublicHostname:       configCmd.GetDefaultPublicHostName(ip),
			SSHCommander:         sshCommander,
			OcBinaryPathInsideVM: fmt.Sprintf("%s/oc", minishiftConstants.OcPathInsideVM),
		}

		clusterUpParams := determineClusterUpParameters(clusterUpConfig)
		fmt.Println("-- OpenShift cluster will be configured with ...")
		fmt.Println("   Version:", requestedOpenShiftVersion)

		fmt.Printf("-- Copying oc binary from the OpenShift container image to VM ")
		progressDots := progressdots.New()
		progressDots.Start()
		err = clusterup.CopyOcBinaryFromImageToVM(dockerCommander, minishiftConstants.GetOpenshiftImageName(requestedOpenShiftVersion), minishiftConstants.OcPathInsideVM)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error copying the oc binary to %s: %v", minishiftConstants.OcPathInsideVM, err))
		}
		progressDots.Stop()
		fmt.Println(" OK")

		fmt.Printf("-- Starting OpenShift cluster ")
		progressDots = progressdots.New()
		progressDots.Start()
		out, err := clusterup.ClusterUp(clusterUpConfig, clusterUpParams)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error during 'cluster up' execution: %v", err))
		}
		progressDots.Stop()
		fmt.Printf("\n%s\n", out)

		if !IsOpenShiftRunning(hostVm.Driver) {
			atexit.ExitWithMessage(1, "OpenShift provisioning failed. origin container failed to start.")
		}

		if !isRestart {
			postClusterUp(hostVm, clusterUpConfig)
			exportContainerImages(hostVm.Driver, libMachineClient, requestedOpenShiftVersion)
		}
		if isRestart {
			err = cmdUtil.SetOcContext(minishiftConfig.AllInstancesConfig.ActiveProfile)
			if err != nil {
				fmt.Println(fmt.Sprintf("Could not set oc CLI context for: '%s'", profileActions.GetActiveProfile()))
			}
		}
	}
}

// postClusterUp performs configuration action which only need to be run after an initial provision of OpenShift.
// On subsequent VM restarts these actions can be skipped.
func postClusterUp(hostVm *host.Host, clusterUpConfig *clusterup.ClusterUpConfig) {
	sshCommander := provision.GenericSSHCommander{Driver: hostVm.Driver}
	err := clusterup.PostClusterUp(clusterUpConfig, sshCommander, addon.GetAddOnManager(), &util.RealRunner{})
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error during post cluster up configuration: %v", err))
	}
}

// getRequiredHostDirectories returns a list of directories we need to ensure exist on the VM.
func getRequiredHostDirectories() []string {
	var requiredDirectories []string
	for _, value := range dataDirectory {
		requiredDirectories = append(requiredDirectories, value)
	}
	requiredDirectories = append(requiredDirectories, minishiftConstants.OcPathInsideVM)
	return requiredDirectories
}

func handleProxies() *util.ProxyConfig {
	proxyConfig, err := util.NewProxyConfig(viper.GetString(cmdUtil.HttpProxy), viper.GetString(cmdUtil.HttpsProxy), viper.GetString(configCmd.NoProxyList.Name))

	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	if proxyConfig.IsEnabled() {
		fmt.Println("-- Using proxy for the setup")
		proxyConfig.ApplyToEnvironment()
		dockerEnv = append(dockerEnv, proxyConfig.ProxyConfig()...)
		shellProxyEnv += strings.Join(proxyConfig.ProxyConfig(), " ")

		// It could be that the proxy config is retrieved from the environment. To make sure that
		// proxy settings are properly passed to cluster up we need to explicitly set the values.
		if proxyConfig.HttpProxy() != "" {
			viper.Set(cmdUtil.HttpProxy, proxyConfig.HttpProxy())
			if glog.V(5) {
				fmt.Println(fmt.Sprintf("\tUsing http proxy: %s", proxyConfig.HttpProxy()))
			}
		}

		if proxyConfig.HttpsProxy() != "" {
			viper.Set(cmdUtil.HttpsProxy, proxyConfig.HttpsProxy())
			if glog.V(5) {
				fmt.Println(fmt.Sprintf("\tUsing https proxy: %s", proxyConfig.HttpsProxy()))
			}
		}
		viper.Set(configCmd.NoProxyList.Name, proxyConfig.NoProxy())
	}

	return proxyConfig
}

// getSlice return slice for provided key otherwise nil
func getSlice(key string) []string {
	if viper.IsSet(key) {
		return viper.GetStringSlice(key)
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

func startHost(libMachineClient *libmachine.Client) *host.Host {
	progressDots := progressdots.New()

	// Configuration used for creation/setup of the Virtual Machine
	machineConfig := &cluster.MachineConfig{
		MinikubeISO:         determineIsoUrl(viper.GetString(configCmd.ISOUrl.Name)),
		ISOCacheDir:         state.InstanceDirs.IsoCache,
		Memory:              calculateMemorySize(viper.GetString(configCmd.Memory.Name)),
		CPUs:                viper.GetInt(configCmd.CPUs.Name),
		DiskSize:            calculateDiskSize(viper.GetString(configCmd.DiskSize.Name)),
		VMDriver:            viper.GetString(configCmd.VmDriver.Name),
		DockerEnv:           append(dockerEnv, getSlice(configCmd.DockerEnv.Name)...),
		DockerEngineOpt:     getSlice(configCmd.DockerEngineOpt.Name),
		InsecureRegistry:    determineInsecureRegistry(configCmd.InsecureRegistry.Name),
		RegistryMirror:      getSlice(configCmd.RegistryMirror.Name),
		HostOnlyCIDR:        viper.GetString(configCmd.HostOnlyCIDR.Name),
		HypervVirtualSwitch: viper.GetString(configCmd.HypervVirtualSwitch.Name),
		ShellProxyEnv:       shellProxyEnv,
	}
	minishiftConfig.InstanceConfig.VMDriver = machineConfig.VMDriver
	minishiftConfig.InstanceConfig.Write()

	fmt.Printf(" using '%s' hypervisor ...\n", machineConfig.VMDriver)
	var hostVm *host.Host

	// configuration with these settings only happen on create
	isRestart := cmdUtil.VMExists(libMachineClient, constants.MachineName)
	if !isRestart {
		fmt.Println("-- Minishift VM will be configured with ...")
		fmt.Println("   Memory:   ", units.HumanSize(float64((machineConfig.Memory/units.KiB)*units.GB)))
		fmt.Println("   vCPUs :   ", machineConfig.CPUs)
		fmt.Println("   Disk size:", units.HumanSize(float64(machineConfig.DiskSize*units.MB)))
	}

	configureNetworkSettings()

	cacheMinishiftISO(machineConfig)

	fmt.Print("-- Starting Minishift VM ...")
	progressDots.Start()
	start := func() (err error) {
		hostVm, err = cluster.StartHost(libMachineClient, *machineConfig)
		if err != nil {
			fmt.Print(" FAIL ")
			glog.Errorf("Error starting the VM: %s. Retrying.\n", err)
		}
		return err
	}
	err := util.Retry(3, start)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error starting the VM: %v", err))
	}
	progressDots.Stop()

	fmt.Println(" OK")
	return hostVm
}

func configureNetworkSettings() {
	networkSettings := minishiftNetwork.NetworkSettings{
		Device:    viper.GetString(configCmd.NetworkDevice.Name),
		IPAddress: viper.GetString(configCmd.IPAddress.Name),
		Netmask:   viper.GetString(configCmd.Netmask.Name),
		Gateway:   viper.GetString(configCmd.Gateway.Name),
	}

	nameservers := getSlice(configCmd.NameServers.Name)
	if len(nameservers) > 0 {
		networkSettings.DNS1 = nameservers[0]
	}
	if len(nameservers) > 1 {
		networkSettings.DNS2 = nameservers[1]
	}

	// Configure networking on startup only works on Hyper-V
	if networkSettings.IPAddress != "" {
		minishiftNetwork.ConfigureNetworking(constants.MachineName, networkSettings)
	}
}

func autoMountHostFolders(driver drivers.Driver) {
	hostFolderManager, err := hostfolder.NewManager(minishiftConfig.InstanceConfig, minishiftConfig.AllInstancesConfig)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}
	if isAutoMount() && hostFolderManager.ExistAny() {
		fmt.Println("-- Mounting host folders")
		hostFolderManager.MountAll(driver)
	}
}

func isNoProvision() bool {
	return viper.GetBool(configCmd.NoProvision.Name)
}

func isAutoMount() bool {
	return viper.GetBool(configCmd.HostFoldersAutoMount.Name)
}

func addActiveProfileInformation() {
	if constants.ProfileName != profileActions.GetActiveProfile() {
		fmt.Println(fmt.Sprintf("-- Switching active profile to '%s'", constants.ProfileName))
		err := profileActions.SetActiveProfile(constants.ProfileName)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
	}
}

func importContainerImages(driver drivers.Driver, api libmachine.API, openShiftVersion string) {
	if !viper.GetBool(configCmd.ImageCaching.Name) {
		return
	}

	images := viper.GetStringSlice(configCmd.CacheImages.Name)
	for _, coreImage := range image.GetOpenShiftImageNames(openShiftVersion) {
		if !stringUtils.Contains(images, coreImage) {
			images = append(images, coreImage)
		}
	}

	envMap, err := cluster.GetHostDockerEnv(api)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error determining Docker settings for image import: %v", err))
	}

	handler := getImageHandler(driver, envMap)
	config := &image.ImageCacheConfig{
		HostCacheDir: state.InstanceDirs.ImageCache,
		CachedImages: images,
		Out:          os.Stdout,
	}
	_, err = handler.ImportImages(config)
	if err != nil {
		fmt.Println(fmt.Sprintf("  WARN: At least one image could not be imported. Error: %s ", err.Error()))
	}
}

func getImageHandler(driver drivers.Driver, envMap map[string]string) image.ImageHandler {
	handler, err := image.NewOciImageHandler(driver, envMap)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Unable to create image handler: %v", err))
	}

	return handler
}

// exportContainerImages exports the OpenShift images in a background process (by calling 'minishift image export')
func exportContainerImages(driver drivers.Driver, api libmachine.API, version string) {
	if !viper.GetBool(configCmd.ImageCaching.Name) {
		return
	}

	images := viper.GetStringSlice(configCmd.CacheImages.Name)
	for _, coreImage := range image.GetOpenShiftImageNames(version) {
		if !stringUtils.Contains(images, coreImage) {
			images = append(images, coreImage)
		}
	}

	envMap, err := cluster.GetHostDockerEnv(api)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error determining Docker settings for image import: %v", err))
	}

	handler := getImageHandler(driver, envMap)
	config := &image.ImageCacheConfig{
		HostCacheDir: state.InstanceDirs.ImageCache,
		CachedImages: images,
	}

	if handler.AreImagesCached(config) {
		return
	}

	exportCmd, err := image.CreateExportCommand(version, constants.ProfileName, images)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating export command: %v", err))
	}

	err = exportCmd.Start()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error during export: %v", err))
	}
	fmt.Println(fmt.Sprintf("-- Exporting of OpenShift images is occuring in background process with pid %d.", exportCmd.Process.Pid))
}

func calculateMemorySize(memorySize string) int {
	if stringUtils.HasOnlyNumbers(memorySize) {
		memorySize += "MB"
	}

	// err := minishiftConfig.IsValidMemorySize(configCmd.Memory.Name, humanReadableSize)
	size, err := units.RAMInBytes(memorySize)
	if err != nil {
		fmt.Println()
		atexit.ExitWithMessage(1, fmt.Sprintf("Memory size is not valid: %v", err))
	}

	return int(size / units.MiB)
}

func calculateDiskSize(humanReadableSize string) int {
	if stringUtils.HasOnlyNumbers(humanReadableSize) {
		humanReadableSize += "MB"
	}

	// err := minishiftConfig.IsValidDiskSize(configCmd.DiskSize.Name, humanReadableSize)
	size, err := units.FromHumanSize(humanReadableSize)
	if err != nil {
		fmt.Println()
		atexit.ExitWithMessage(1, fmt.Sprintf("Disk size is not valid: %v", err))
	}

	return int(size / units.MB)
}

func determineIsoUrl(iso string) string {
	isoNotSpecified := ""

	switch strings.ToLower(iso) {
	case minishiftConstants.B2dIsoAlias, isoNotSpecified:
		iso = constants.DefaultB2dIsoUrl
	case minishiftConstants.CentOsIsoAlias:
		iso = constants.DefaultCentOsIsoUrl
	case minishiftConstants.MinikubeIsoAlias:
		fmt.Printf("\n\nNote: minikube ISO will be depricated in next release, please use b2d or centos ISO.\n\n")
		iso = constants.DefaultMinikubeIsoURL
	default:
		if !(govalidator.IsURL(iso) || strings.HasPrefix(iso, "file:")) {
			fmt.Println()
			atexit.ExitWithMessage(1, unsupportedIsoUrlFormat)
		}
	}

	return iso
}

// initStartFlags creates the CLI flags which needs to be passed on to 'libmachine'
func initStartFlags() *flag.FlagSet {
	startFlagSet := flag.NewFlagSet(commandName, flag.ContinueOnError)

	startFlagSet.String(configCmd.VmDriver.Name, constants.DefaultVMDriver, fmt.Sprintf("The driver to use for the Minishift VM. Possible values: %v", constants.SupportedVMDrivers))
	startFlagSet.Int(configCmd.CPUs.Name, constants.DefaultCPUS, "Number of CPU cores to allocate to the Minishift VM.")
	startFlagSet.String(configCmd.Memory.Name, constants.DefaultMemory, "Amount of RAM to allocate to the Minishift VM. Use the format <size><unit>, where unit = MB or GB.")
	startFlagSet.String(configCmd.DiskSize.Name, constants.DefaultDiskSize, "Disk size to allocate to the Minishift VM. Use the format <size><unit>, where unit = MB or GB.")
	startFlagSet.String(configCmd.HostOnlyCIDR.Name, "192.168.99.1/24", "The CIDR to be used for the minishift VM. (Only supported with VirtualBox driver.)")
	startFlagSet.Bool(configCmd.SkipPreflightChecks.Name, false, "Skip the startup checks.")
	startFlagSet.AddFlag(dockerEnvFlag)
	startFlagSet.AddFlag(dockerEngineOptFlag)
	startFlagSet.AddFlag(insecureRegistryFlag)
	startFlagSet.AddFlag(registryMirrorFlag)
	startFlagSet.AddFlag(cmdUtil.AddOnEnvFlag)
	startFlagSet.String(configCmd.OpenshiftVersion.Name, version.GetOpenShiftVersion(), fmt.Sprintf("The OpenShift version to run, eg. %s", version.GetOpenShiftVersion()))

	if runtime.GOOS == "windows" {
		startFlagSet.String(configCmd.NetworkDevice.Name, "", "Specify the network device to use for the IP address. Ignored if no IP address specified (Hyper-V only)")
		startFlagSet.String(configCmd.IPAddress.Name, "", "Specify IP address to assign to the instance (Hyper-V only)")
		startFlagSet.String(configCmd.Netmask.Name, "", "Specify netmask to use for the IP address. Ignored if no IP address specified (Hyper-V only)")
		startFlagSet.String(configCmd.Gateway.Name, "", "Specify gateway to use for the instance. Ignored if no IP address specified (Hyper-V only)")
		startFlagSet.String(configCmd.HypervVirtualSwitch.Name, "", "Specify which Virtual Switch to use for the instance (Hyper-V only)")
	}
	startFlagSet.AddFlag(nameServersFlag)

	if minishiftConfig.EnableExperimental {
		startFlagSet.Bool(configCmd.NoProvision.Name, false, "Do not provision the VM with OpenShift (experimental)")
		startFlagSet.String(configCmd.ISOUrl.Name, minishiftConstants.B2dIsoAlias, "Location of the minishift ISO. Can be an URL, file URI or one of the following short names: [b2d centos minikube].")
	} else {
		startFlagSet.String(configCmd.ISOUrl.Name, minishiftConstants.B2dIsoAlias, "Location of the minishift ISO. Can be an URL, file URI or one of the following short names: [b2d centos].")
	}

	return startFlagSet
}

// initClusterUpFlags creates the CLI flags which needs to be passed on to 'oc cluster up'
func initClusterUpFlags() *flag.FlagSet {
	clusterUpFlagSet := flag.NewFlagSet(commandName, flag.ContinueOnError)

	clusterUpFlagSet.Bool(configCmd.SkipRegistryCheck.Name, false, "Skip the Docker daemon registry check.")
	clusterUpFlagSet.String(configCmd.PublicHostname.Name, "", "Public hostname of the OpenShift cluster.")
	clusterUpFlagSet.String(configCmd.RoutingSuffix.Name, "", "Default suffix for the server routes.")
	clusterUpFlagSet.Int(configCmd.ServerLogLevel.Name, 0, "Log level for the OpenShift server.")
	clusterUpFlagSet.String(configCmd.NoProxyList.Name, "", "List of hosts or subnets for which no proxy should be used.")
	clusterUpFlagSet.AddFlag(cmdUtil.HttpProxyFlag)
	clusterUpFlagSet.AddFlag(cmdUtil.HttpsProxyFlag)

	if minishiftConfig.EnableExperimental {
		clusterUpFlagSet.String(configCmd.ExtraClusterUpFlags.Name, "", "Specify optional flags for use with 'cluster up' (unsupported)")
	}

	return clusterUpFlagSet
}

// initSubscriptionManagerFlags create the CLI flags which are needed for VM registration
func initSubscriptionManagerFlags() *flag.FlagSet {
	subscriptionManagerFlagSet := flag.NewFlagSet(commandName, flag.ContinueOnError)
	subscriptionManagerFlagSet.String(configCmd.Username.Name, "", "Username for the virtual machine registration.")
	subscriptionManagerFlagSet.String(configCmd.Password.Name, "", "Password for the virtual machine registration.")
	subscriptionManagerFlagSet.Bool(configCmd.SkipRegistration.Name, false, "Skip the virtual machine registration.")

	return subscriptionManagerFlagSet
}

// determineClusterUpParameters returns a map of flag names and values for the cluster up call.
func determineClusterUpParameters(config *clusterup.ClusterUpConfig) map[string]string {
	clusterUpParams := make(map[string]string)

	// Set default value for host config, data and volumes
	viper.Set(configCmd.RoutingSuffix.Name, config.RoutingSuffix)
	viper.Set(configCmd.PublicHostname.Name, config.PublicHostname)

	clusterUpFlagSet.VisitAll(func(flag *flag.Flag) {
		if viper.IsSet(flag.Name) {
			value := viper.GetString(flag.Name)
			key := flag.Name
			clusterUpParams[key] = value
		}
	})

	// This will only required to work with openshift < 3.10
	for key, value := range dataDirectory {
		clusterUpParams[key] = value
	}

	return clusterUpParams
}

func ensureNotRunning(client *libmachine.Client, machineName string) {
	if !cmdUtil.VMExists(client, machineName) {
		return
	}

	hostVm, err := client.Load(constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	if cmdUtil.IsHostRunning(hostVm.Driver) {
		atexit.ExitWithMessage(0, fmt.Sprintf("The '%s' VM is already running.", machineName))
	}
}

// Make sure the version actually has a 'v' prefix. See https://github.com/minishift/minishift/issues/410
func addVersionPrefixToOpenshiftVersion() {
	requestedVersion := viper.GetString(configCmd.OpenshiftVersion.Name)
	if !strings.HasPrefix(requestedVersion, constants.VersionPrefix) {
		requestedVersion = constants.VersionPrefix + requestedVersion
		// this will make sure the right version is set in case the version comes from config file
		viper.Set(configCmd.OpenshiftVersion.Name, requestedVersion)

		// if the version was specified via the CLI we need to update the flag value
		startCmd.Flags().Lookup(configCmd.OpenshiftVersion.Name).Value.Set(requestedVersion)
	}
}

func setSubscriptionManagerParameters() {
	registrationUtil.SkipRegistration = viper.GetBool(configCmd.SkipRegistration.Name)
	minishiftCluster.RegistrationParameters.Username = viper.GetString(configCmd.Username.Name)
	minishiftCluster.RegistrationParameters.Password = viper.GetString(configCmd.Password.Name)
	minishiftCluster.RegistrationParameters.IsTtySupported = util.IsTtySupported()
	minishiftCluster.RegistrationParameters.GetUsernameInteractive = getUsernameInteractive
	minishiftCluster.RegistrationParameters.GetPasswordInteractive = getPasswordInteractive
}

func getUsernameInteractive(message string) string {
	return util.ReadInputFromStdin(message)
}

func getPasswordInteractive(message string) string {
	return util.ReadPasswordFromStdin(message)
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

func cacheMinishiftISO(config *cluster.MachineConfig) {
	if config.ShouldCacheMinikubeISO() {
		if err := config.CacheMinikubeISOFromURL(); err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error caching the ISO: %s", err.Error()))
		}
	}
}

// if skip-startup-checks set to true then return true and skip preflight checks
func shouldPreflightChecksBeSkipped() bool {
	return viper.GetBool(configCmd.SkipPreflightChecks.Name)
}
