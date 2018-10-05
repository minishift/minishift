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
	"github.com/minishift/minishift/pkg/minishift/timezone"
	"os"
	"runtime"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/docker/go-units"
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
	"github.com/minishift/minishift/pkg/minikube/sshutil"
	minishiftCluster "github.com/minishift/minishift/pkg/minishift/cluster"
	"github.com/minishift/minishift/pkg/minishift/clusterup"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/minishift/minishift/pkg/minishift/docker/image"
	"github.com/minishift/minishift/pkg/minishift/hostfolder"
	minishiftNetwork "github.com/minishift/minishift/pkg/minishift/network"
	minishiftProxy "github.com/minishift/minishift/pkg/minishift/network/proxy"
	"github.com/minishift/minishift/pkg/minishift/openshift"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/minishift/provisioner"
	"github.com/minishift/minishift/pkg/minishift/remotehost"
	"github.com/minishift/minishift/pkg/minishift/systemtray"
	minishiftTLS "github.com/minishift/minishift/pkg/minishift/tls"
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
	genericDriver           = "generic"
	dockerbridgeSubnetCmd   = `docker network inspect -f "{{range .IPAM.Config }}{{ .Subnet }}{{end}}" bridge`
)

var (
	dockerEnv               []string
	openShiftEnv            []string
	shellProxyEnv           util.ProxyConfig
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
	// Set the base dir for v3.10.0
	baseDirectory = minishiftConstants.BaseDirInsideInstance

	// clusterUpFlagSet contains the command line switches which needs to be passed on to 'cluster up'
	clusterUpFlagSet *flag.FlagSet
	startFlagSet     *flag.FlagSet

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
	startFlagSet = initStartFlags()
	startCmd.Flags().AddFlagSet(startFlagSet)
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

	if viper.GetString(configCmd.VmDriver.Name) == genericDriver {
		cmdUtil.ValidateGenericDriverFlags(viper.GetString(configCmd.RemoteIPAddress.Name),
			viper.GetString(configCmd.RemoteSSHUser.Name),
			viper.GetString(configCmd.SSHKeyToConnectRemote.Name))
	}

	ensureNotRunning(libMachineClient, constants.MachineName)
	addVersionPrefixToOpenshiftVersion()

	// to determine whether we need to run post cluster up actions,
	// we need to determine whether this is a restart prior to potentially creating a new VM
	isRestart := cmdUtil.VMExists(libMachineClient, constants.MachineName)

	// create and handle proxy config for local environment
	proxyConfig := handleProxyConfig()

	// Get proper OpenShift version
	requestedOpenShiftVersion, err := cmdUtil.GetOpenShiftReleaseVersion()
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting OpenShift version: %v", err))
	}

	// preflight check (before start)
	if viper.GetString(configCmd.VmDriver.Name) != genericDriver {
		preflightChecksBeforeStartingHost()
	}

	// Populate start flags to viper config if save-start-flags true in config file
	if viper.GetBool(configCmd.SaveStartFlags.Name) {
		populateStartFlagsToViperConfig()
	}

	// Cache OC binary before starting the VM and perform oc command option check
	ocPath = cmdUtil.CacheOc(requestedOpenShiftVersion)
	preflightChecksForArtifacts()

	setSubscriptionManagerParameters()

	fmt.Print("-- Starting the OpenShift cluster")

	hostVm := startHost(libMachineClient)
	if !isRestart {
		minishiftConfig.InstanceStateConfig.TimeZone = viper.GetString(configCmd.TimeZone.Name)
		minishiftConfig.InstanceStateConfig.Write()
	}
	timezone.SetTimeZone(hostVm)
	registrationUtil.RegisterHost(libMachineClient)

	// Forcibly set nameservers when configured
	minishiftNetwork.AddNameserversToInstance(hostVm.Driver, getSlice(configCmd.NameServers.Name))
	// to support intermediate proxy
	minishiftTLS.SetCACertificate(hostVm.Driver)

	ip, _ := hostVm.Driver.GetIP()
	localProxy := viper.GetBool(configCmd.LocalProxy.Name)
	// This is a hack/workaround as the actual values are modified in `cluster.go` for the VM
	if proxyConfig.IsEnabled() {
		// Once we know the IP, we need to make sure it is not proxied in a proxy environment.
		// In addition, we also add the host interface's IP to NoProxy so that
		// we can reach the host machine. This is useful when accessing
		// services running on the host machine.
		hostip, _ := minishiftNetwork.DetermineHostIP(hostVm.Driver)

		if localProxy {
			minishiftNetwork.AddHostEntryToInstance(hostVm.Driver, "localproxy", hostip)
			localProxyAddr := fmt.Sprintf("%s:%s", hostip, "3128")
			proxyConfig.OverrideHttpProxy(localProxyAddr)
			proxyConfig.OverrideHttpsProxy(localProxyAddr)
		}

		proxyConfig.AddNoProxy(ip)
		proxyConfig.AddNoProxy(hostip)
		proxyConfig.ApplyToEnvironment()
	}

	// preflight checks (after start)
	if viper.GetString(configCmd.VmDriver.Name) != genericDriver {
		preflightChecksAfterStartingHost(hostVm.Driver)
	}

	// Adding active profile information to all instance config
	addActiveProfileInformation()

	applyDockerEnvToProcessEnv(libMachineClient)

	err = clusterup.EnsureHostDirectoriesExist(hostVm, getRequiredHostDirectories())
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating required host directories: %v", err))
	}

	autoMountHostFolders(hostVm.Driver)

	// start the minishift system tray
	err = startTray()
	if err != nil {
		fmt.Println(err)
	}

	if !isNoProvision() {
		if !isRestart {
			importContainerImages(hostVm.Driver, libMachineClient, requestedOpenShiftVersion)
		}

		sshCommander := provision.GenericSSHCommander{Driver: hostVm.Driver}
		dockerCommander := docker.NewVmDockerCommander(sshCommander)
		dockerbridgeSubnet, err := sshCommander.SSHCommand(dockerbridgeSubnetCmd)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

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
			SshUser:              sshCommander.Driver.GetSSHUsername(),
		}

		clusterUpParams := determineClusterUpParameters(clusterUpConfig, strings.TrimSpace(dockerbridgeSubnet))
		fmt.Println("-- OpenShift cluster will be configured with ...")
		fmt.Println("   Version:", requestedOpenShiftVersion)

		err = cmdUtil.PullOpenshiftImageAndCopyOcBinary(dockerCommander, requestedOpenShiftVersion)
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		fmt.Printf("-- Starting OpenShift cluster ")
		progressDots := progressdots.New()
		progressDots.Start()

		if localProxy {
			// workaround for non-persistence of proxy config
			clusterUpParams["http-proxy"] = proxyConfig.HttpProxy()
			clusterUpParams["https-proxy"] = proxyConfig.HttpsProxy()
		}

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
				atexit.ExitWithMessage(1, fmt.Sprintf("Could not set oc CLI context for '%s' profile: %v", profileActions.GetActiveProfile(), err))
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
	requiredDirectories = append(requiredDirectories, baseDirectory)
	requiredDirectories = append(requiredDirectories, minishiftConstants.OcPathInsideVM)
	return requiredDirectories
}

func handleProxyConfig() *util.ProxyConfig {
	httpProxy := viper.GetString(cmdUtil.HttpProxy)
	httpsProxy := viper.GetString(cmdUtil.HttpsProxy)
	noProxy := viper.GetString(configCmd.NoProxyList.Name)

	localProxy := viper.GetBool(configCmd.LocalProxy.Name)
	if localProxy {
		fmt.Println("-- Starting local proxy")
		err := minishiftProxy.EnsureProxyDaemonRunning()
		httpProxy = "http://localproxy:3128"
		httpsProxy = "http://localproxy:3128"
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		minishiftNetwork.OverrideInsecureSkipVerifyForLocalConnections(true)
		minishiftNetwork.OverrideProxyForLocalConnections("http://localhost:3128")
	}

	proxyConfig, err := util.NewProxyConfig(httpProxy, httpsProxy, noProxy)

	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	if proxyConfig.IsEnabled() {
		fmt.Println("-- Using proxy for the setup")
		proxyConfig.ApplyToEnvironment()
		dockerEnv = append(dockerEnv, proxyConfig.ProxyConfig()...)
		shellProxyEnv = *proxyConfig

		// It could be that the proxy config is retrieved from the environment. To make sure that
		// proxy settings are properly passed to cluster up we need to explicitly set the values.
		if proxyConfig.HttpProxy() != "" && !localProxy {
			// workaround - as these should never be persisted here
			viper.Set(cmdUtil.HttpProxy, proxyConfig.HttpProxy())
			if glog.V(5) {
				fmt.Println(fmt.Sprintf("\tUsing http proxy: %s", proxyConfig.HttpProxy()))
			}
		}

		if proxyConfig.HttpsProxy() != "" && !localProxy {
			// workaround - as these should never be persisted here
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
	// Configuration used for creation/setup of the Virtual Machine
	machineConfig := &cluster.MachineConfig{
		MinikubeISO:           determineIsoUrl(viper.GetString(configCmd.ISOUrl.Name)),
		ISOCacheDir:           state.InstanceDirs.IsoCache,
		Memory:                calculateMemorySize(viper.GetString(configCmd.Memory.Name)),
		CPUs:                  viper.GetInt(configCmd.CPUs.Name),
		DiskSize:              calculateDiskSize(viper.GetString(configCmd.DiskSize.Name)),
		VMDriver:              viper.GetString(configCmd.VmDriver.Name),
		DockerEnv:             append(dockerEnv, getSlice(configCmd.DockerEnv.Name)...),
		DockerEngineOpt:       getSlice(configCmd.DockerEngineOpt.Name),
		InsecureRegistry:      determineInsecureRegistry(configCmd.InsecureRegistry.Name),
		RegistryMirror:        getSlice(configCmd.RegistryMirror.Name),
		HostOnlyCIDR:          viper.GetString(configCmd.HostOnlyCIDR.Name),
		HypervVirtualSwitch:   viper.GetString(configCmd.HypervVirtualSwitch.Name),
		ShellProxyEnv:         shellProxyEnv,
		RemoteIPAddress:       viper.GetString(configCmd.RemoteIPAddress.Name),
		RemoteSSHUser:         viper.GetString(configCmd.RemoteSSHUser.Name),
		SSHKeyToConnectRemote: viper.GetString(configCmd.SSHKeyToConnectRemote.Name),
		UsingLocalProxy:       viper.GetBool(configCmd.LocalProxy.Name),
	}
	minishiftConfig.InstanceStateConfig.VMDriver = machineConfig.VMDriver
	minishiftConfig.InstanceStateConfig.Write()

	fmt.Printf(" using '%s' hypervisor ...\n", machineConfig.VMDriver)
	var hostVm *host.Host

	if machineConfig.VMDriver != genericDriver {
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
	} else {
		s, err := sshutil.NewRawSSHClient(machineConfig.RemoteIPAddress, machineConfig.SSHKeyToConnectRemote, machineConfig.RemoteSSHUser)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error creating ssh client: %v", err))
		}
		fmt.Printf("-- Preparing Remote Machine ...")
		progressDots := progressdots.New()
		progressDots.Start()
		if err := remotehost.PrepareRemoteMachine(s); err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
		progressDots.Stop()
		fmt.Println(" OK")
		fmt.Print("-- Starting to provision the remote machine ...")
	}
	progressDots := progressdots.New()
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

	images := minishiftConfig.InstanceConfig.CacheImages
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

	images := minishiftConfig.InstanceConfig.CacheImages
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
	startFlagSet.String(configCmd.OpenshiftVersion.Name, version.GetOpenShiftVersion(), fmt.Sprintf("The OpenShift version to run, eg. latest or %s", version.GetOpenShiftVersion()))

	startFlagSet.String(configCmd.RemoteIPAddress.Name, "", "IP address of the remote machine to provision OpenShift on")
	startFlagSet.String(configCmd.RemoteSSHUser.Name, "", "The username of the remote machine to provision OpenShift on")
	startFlagSet.String(configCmd.SSHKeyToConnectRemote.Name, "", "SSH private key location on the host to connect remote machine")
	startFlagSet.String(configCmd.ISOUrl.Name, minishiftConstants.CentOsIsoAlias, "Location of the minishift ISO. Can be a URL, file URI or one of the following short names: [centos b2d].")
	startFlagSet.String(configCmd.TimeZone.Name, constants.DefaultTimeZone, "TimeZone for Minishift VM")

	startFlagSet.AddFlag(dockerEnvFlag)
	startFlagSet.AddFlag(dockerEngineOptFlag)
	startFlagSet.AddFlag(insecureRegistryFlag)
	startFlagSet.AddFlag(registryMirrorFlag)
	startFlagSet.AddFlag(cmdUtil.AddOnEnvFlag)

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
	clusterUpFlagSet.String(configCmd.ImageName.Name, "", "Specify the images to use for OpenShift")
	clusterUpFlagSet.AddFlag(cmdUtil.HttpProxyFlag)
	clusterUpFlagSet.AddFlag(cmdUtil.HttpsProxyFlag)
	// This is hidden because we don't want our users to use this flag
	// we are setting it to openshift/origin-${component}:<user_provided_version> as default
	// It is used for testing purpose from CDK/minishift QE
	clusterUpFlagSet.MarkHidden(configCmd.ImageName.Name)

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

// Get config from the start flags and set it to viper config so that in
// stop start case user don't need to remember the start flags.
func populateStartFlagsToViperConfig() {
	startFlagSet.AddFlag(cmdUtil.HttpProxyFlag)
	startFlagSet.AddFlag(cmdUtil.HttpsProxyFlag)
	startFlagSet.VisitAll(func(flag *flag.Flag) {
		if viper.IsSet(flag.Name) {
			switch value := viper.Get(flag.Name).(type) {
			case string:
				if err := configCmd.Set(flag.Name, viper.GetString(flag.Name), false); err != nil {
					atexit.ExitWithMessage(1, fmt.Sprintf("Not able to populate %s flag with %s value", flag.Name, value))
				}
			case int:
				if err := configCmd.Set(flag.Name, viper.GetString(flag.Name), false); err != nil {
					atexit.ExitWithMessage(1, fmt.Sprintf("Not able to populate %s flag with %d value", flag.Name, value))
				}
			case []interface{}:
				if err := configCmd.Set(flag.Name, strings.Join(viper.GetStringSlice(flag.Name), ","), false); err != nil {
					atexit.ExitWithMessage(1, fmt.Sprintf("Not able to populate %s flag with %#v value", flag.Name, value))
				}
			case bool:
				if err := configCmd.Set(flag.Name, viper.GetString(flag.Name), false); err != nil {
					atexit.ExitWithMessage(1, fmt.Sprintf("Not able to populate %s flag with %#v value", flag.Name, value))
				}
			}
		}
	})
}

// determineClusterUpParameters returns a map of flag names and values for the cluster up call.
func determineClusterUpParameters(config *clusterup.ClusterUpConfig, DockerbridgeSubnet string) map[string]string {
	clusterUpParams := make(map[string]string)
	// Set default value for base config for 3.10
	clusterUpParams["base-dir"] = baseDirectory
	if viper.GetString(configCmd.ImageName.Name) == "" {
		imagetag := fmt.Sprintf("'%s:%s'", minishiftConstants.ImageNameForClusterUpImageFlag, config.OpenShiftVersion)
		viper.Set(configCmd.ImageName.Name, imagetag)
	}
	// Add docker bridge subnet to no-proxy before passing to oc cluster up
	if viper.GetString(configCmd.NoProxyList.Name) != "" {
		viper.Set(configCmd.NoProxyList.Name, fmt.Sprintf("%s,%s", DockerbridgeSubnet, viper.GetString(configCmd.NoProxyList.Name)))
	}

	viper.Set(configCmd.RoutingSuffix.Name, config.RoutingSuffix)
	viper.Set(configCmd.PublicHostname.Name, config.PublicHostname)
	clusterUpFlagSet.VisitAll(func(flag *flag.Flag) {
		if viper.IsSet(flag.Name) {
			value := viper.GetString(flag.Name)
			key := flag.Name
			clusterUpParams[key] = value
		}
	})

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

	if cmdUtil.IsHostRunning(hostVm.Driver) && hostVm.DriverName != "generic" {
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

func startTray() error {
	if runtime.GOOS != "linux" {
		minishiftTray := systemtray.NewMinishiftTray(minishiftConfig.AllInstancesConfig)
		err := minishiftTray.EnsureRunning()
		if err != nil {
			return err
		}
	}
	return nil
}
