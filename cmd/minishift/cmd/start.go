/*
Copyright (C) 2016 Red Hat, Inc.

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

	units "github.com/docker/go-units"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/cache"
	"github.com/minishift/minishift/pkg/minishift/clusterup"
	"github.com/minishift/minishift/pkg/minishift/provisioner"
	minishiftUtil "github.com/minishift/minishift/pkg/minishift/util"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/minishift/minishift/pkg/version"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	commandName = "start"

	// minishift
	isoURL                = "iso-url"
	memory                = "memory"
	cpus                  = "cpus"
	humanReadableDiskSize = "disk-size"
	vmDriver              = "vm-driver"
	openshiftVersion      = "openshift-version"
	hostOnlyCIDR          = "host-only-cidr"

	// cluster up
	skipRegistryCheck = "skip-registry-check"
	publicHostname    = "public-hostname"
	routingSuffix     = "routing-suffix"
	hostConfigDir     = "host-config-dir"
	hostVolumesDir    = "host-volumes-dir"
	hostDataDir       = "host-data-dir"
	forwardPorts      = "forward-ports"
	serverLogLevel    = "server-loglevel"
	openshiftEnv      = "openshift-env"
	metrics           = "metrics"

	// Setting proxy
	httpProxy   = "http-proxy"
	httpsProxy  = "https-proxy"
	noProxyList = "no-proxy"
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
)

// startFlagSet contains the minishift specific command line switches
var startFlagSet = flag.NewFlagSet(commandName, flag.ContinueOnError)

// clusterUpFlagSet contains the command line switches which needs to be passed on to 'cluster up'
var clusterUpFlagSet = flag.NewFlagSet(commandName, flag.ContinueOnError)

// proxyFlagSet contains the command line switches for proxy
var proxyFlagSet = flag.NewFlagSet(commandName, flag.ContinueOnError)

// minishiftToClusterUp is a mapping between falg names used in minishift CLI and flag name as passed to 'cluster up'
var minishiftToClusterUp = map[string]string{
	"openshift-env": "env",
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

	validateProxyArgs()

	setDockerProxy()
	setOcProxy()
	setShellProxy()
	SetRegistrationProxyParameters()

	config := cluster.MachineConfig{
		MinikubeISO:      viper.GetString(isoURL),
		Memory:           viper.GetInt(memory),
		CPUs:             viper.GetInt(cpus),
		DiskSize:         calculateDiskSizeInMB(viper.GetString(humanReadableDiskSize)),
		VMDriver:         viper.GetString(vmDriver),
		DockerEnv:        dockerEnv,
		InsecureRegistry: insecureRegistry,
		RegistryMirror:   registryMirror,
		HostOnlyCIDR:     viper.GetString(hostOnlyCIDR),
		OpenShiftVersion: viper.GetString(openshiftVersion),
	}

	fmt.Printf("Starting local OpenShift cluster using '%s' hypervisor...\n", config.VMDriver)

	var host *host.Host
	start := func() (err error) {
		host, err = cluster.StartHost(libMachineClient, config)
		if err != nil {
			glog.Errorf("Error starting the VM: %s. Retrying.\n", err)
		}
		return err
	}
	err := util.Retry(3, start)
	if err != nil {
		glog.Errorln("Error starting the VM: ", err)
		atexit.Exit(1)
	}

	// Set Proxy to as Shell Env
	if viper.IsSet("http-proxy") || viper.IsSet("https-proxy") {
		if err := minishiftUtil.SetProxyToShellEnv(host, shellProxyEnv); err != nil {
			fmt.Printf("Error setting proxy to VM: %s", err)
			atexit.Exit(1)
		}
	}

	// Making sure the required Docker environment variables are set to make 'cluster up' work
	envMap, err := cluster.GetHostDockerEnv(libMachineClient)
	for k, v := range envMap {
		os.Setenv(k, v)
	}

	//Create the host directories if not present
	hostDirs := []string{viper.GetString(hostConfigDir), viper.GetString(hostDataDir), viper.GetString(hostVolumesDir)}
	err = clusterup.EnsureHostDirectoriesExist(libMachineClient, hostDirs)
	if err != nil {
		glog.Errorln("Error creating required host directories: ", err)
		atexit.Exit(1)
	}

	clusterUp(&config)
}

// Set Docker Proxy
func setDockerProxy() {
	if viper.IsSet(httpProxy) {
		dockerEnv = append(dockerEnv, fmt.Sprintf("HTTP_PROXY=%s", viper.GetString(httpProxy)))
	}
	if viper.IsSet(httpsProxy) {
		dockerEnv = append(dockerEnv, fmt.Sprintf("HTTPS_PROXY=%s", viper.GetString(httpsProxy)))
	}
	if viper.IsSet(noProxyList) {
		dockerEnv = append(dockerEnv, fmt.Sprintf("NO_PROXY=%s,%s", updateNoProxyForDocker(),
			viper.GetString(noProxyList)))
	} else if viper.IsSet(httpProxy) || viper.IsSet(httpsProxy) {
		dockerEnv = append(dockerEnv, fmt.Sprintf("NO_PROXY=%s", updateNoProxyForDocker()))
	}

}

// Set OpenShiftProxy
func setOcProxy() {
	if viper.IsSet(httpProxy) {
		clusterUpFlagSet.String(httpProxy, viper.GetString(httpProxy), "HTTP proxy to use for master and builds")
	}
	if viper.IsSet(httpsProxy) {
		clusterUpFlagSet.String(httpsProxy, viper.GetString(httpsProxy), "HTTPS proxy to use for master and builds")
	}
	if viper.IsSet(noProxyList) {
		clusterUpFlagSet.String(noProxyList, viper.GetString(noProxyList), "List of hosts or subnets for which a proxy shouldn't be use")
	}
}

// Set shell proxy
func setShellProxy() {
	if viper.IsSet(httpProxy) {
		shellProxyEnv += fmt.Sprintf("http_proxy=%s ", viper.GetString(httpProxy))
	}
	if viper.IsSet(httpsProxy) {
		shellProxyEnv += fmt.Sprintf("https_proxy=%s ", viper.GetString(httpsProxy))
	}
	if viper.IsSet(noProxyList) {
		shellProxyEnv += fmt.Sprintf("no_proxy=%s", viper.GetString(noProxyList))
	} else if viper.IsSet(httpProxy) || viper.IsSet(httpsProxy) {
		dockerEnv = append(dockerEnv, fmt.Sprintf("NO_PROXY=%s", updateNoProxyForDocker()))
	}
}

// update default no-proxy for docker
func updateNoProxyForDocker() string {
	return "localhost,127.0.0.1,172.30.1.1"
}

//Update RegistrationProxyParameters with proxy information
func SetRegistrationProxyParameters() {
	var proxyUri string

	if viper.IsSet("https-proxy") {
		proxyUri = viper.GetString(httpsProxy)
	} else if viper.IsSet("http-proxy") {
		proxyUri = viper.GetString(httpProxy)
	}

	if proxyUri != "" {
		server, serverPort, user, password, err := minishiftUtil.ParseProxyUri(proxyUri)
		if err != nil {
			glog.Errorf("Not able to parse the proxy URI: %s\n", err)
			atexit.Exit(1)
		}

		if server != "" {
			cluster.RegistrationParameters.ProxyServer = server
		}
		if serverPort != "" {
			cluster.RegistrationParameters.ProxyServerPort = serverPort
		}

		if user != "" {
			cluster.RegistrationParameters.ProxyUsername = user
		}

		if password != "" {
			cluster.RegistrationParameters.ProxyPassword = password
		}
	}
}

// calculateDiskSizeInMB converts a human specified disk size like "1000MB" or "1GB" and converts it into Megabits
func calculateDiskSizeInMB(humanReadableDiskSize string) int {
	diskSize, err := units.FromHumanSize(humanReadableDiskSize)
	if err != nil {
		glog.Errorf("Disk size is not valid: %s", err)
	}
	return int(diskSize / units.MB)
}

// init configures the command line options of this command
func init() {
	initStartFlags()
	initClusterUpFlags()
	initProxyFlags()

	startCmd.Flags().AddFlagSet(startFlagSet)
	startCmd.Flags().AddFlagSet(clusterUpFlagSet)
	startCmd.Flags().AddFlagSet(proxyFlagSet)

	viper.BindPFlags(startCmd.Flags())
	RootCmd.AddCommand(startCmd)

	provision.SetDetector(&provisioner.MinishiftProvisionerDetector{Delegate: provision.StandardDetector{}})
}

// initStartFlags creates the CLI flags which needs to be passed on to 'libmachine'
func initStartFlags() {
	startFlagSet.String(isoURL, constants.DefaultIsoUrl, "Location of the minishift ISO.")
	startFlagSet.String(vmDriver, constants.DefaultVMDriver, fmt.Sprintf("The driver to use for the Minishift VM. Possible values: %v", constants.SupportedVMDrivers))
	startFlagSet.Int(memory, constants.DefaultMemory, "Amount of RAM to allocate to the Minishift VM.")
	startFlagSet.Int(cpus, constants.DefaultCPUS, "Number of CPU cores to allocate to the Minishift VM.")
	startFlagSet.String(humanReadableDiskSize, constants.DefaultDiskSize, "Disk size to allocate to the Minishift VM. Use the format <size><unit>, where unit = b, k, m or g.")
	startFlagSet.String(hostOnlyCIDR, "192.168.99.1/24", "The CIDR to be used for the minishift VM. (Only supported with VirtualBox driver.)")
	startFlagSet.StringArrayVar(&dockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. Use the format <key>=<value>.")
	startFlagSet.StringSliceVar(&insecureRegistry, "insecure-registry", []string{"172.30.0.0/16"}, "Non-secure Docker registries to pass to the Docker daemon.")
	startFlagSet.StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon.")
	startFlagSet.String(openshiftVersion, version.GetOpenShiftVersion(), fmt.Sprintf("The OpenShift version to run, eg. %s", version.GetOpenShiftVersion()))
}

// initClusterUpFlags creates the CLI flags which needs to be passed on to 'oc cluster up'
func initClusterUpFlags() {
	//clusterUpFlagSet.StringVar(&clusterUpConfig.Image, "image", "openshift/origin", "Specify the images to use for OpenShift")
	clusterUpFlagSet.Bool(skipRegistryCheck, false, "Skip the Docker daemon registry check.")
	clusterUpFlagSet.String(publicHostname, "", "Public hostname of the OpenShift cluster.")
	clusterUpFlagSet.String(routingSuffix, "", "Default suffix for the server routes.")
	clusterUpFlagSet.String(hostConfigDir, hostConfigDirectory, "Location of the OpenShift configuration on the Docker host.")
	clusterUpFlagSet.String(hostVolumesDir, hostVolumesDirectory, "Location of the OpenShift volumes on the Docker host.")
	clusterUpFlagSet.String(hostDataDir, hostDataDirectory, "Location of the OpenShift data on the Docker host. If not specified, etcd data will not be persisted on the host.")
	clusterUpFlagSet.Bool(forwardPorts, false, "Use Docker port forwarding to communicate with the origin container. Requires 'socat' locally.")
	clusterUpFlagSet.Int(serverLogLevel, 0, "Log level for the OpenShift server.")
	clusterUpFlagSet.StringSliceVarP(&openShiftEnv, openshiftEnv, "e", []string{}, "Specify key-value pairs of environment variables to set on the OpenShift container.")
	clusterUpFlagSet.Bool(metrics, false, "Install metrics (experimental)")
}

// initProxyFlags create the CLI flags which needs to be passed for proxy
func initProxyFlags() {
	proxyFlagSet.String(httpProxy, "", "HTTP proxy for virtual machine (In the format of http://<username>:<password>@<proxy_host>:<proxy_port>)")
	proxyFlagSet.String(httpsProxy, "", "HTTPS proxy for virtual machine (In the format of https://<username>:<password>@<proxy_host>:<proxy_port>)")
	proxyFlagSet.String(noProxyList, "", "List of hosts or subnets for which proxy should not be used.")
}

// clusterUp downloads and installs the oc binary in order to run 'cluster up'
func clusterUp(config *cluster.MachineConfig) {
	oc := cache.Oc{
		OpenShiftVersion:  config.OpenShiftVersion,
		MinishiftCacheDir: filepath.Join(constants.Minipath, "cache"),
	}
	err := oc.EnsureIsCached()
	if err != nil {
		glog.Errorln("Error starting the cluster: ", err)
		atexit.Exit(1)
	}

	cmdName := filepath.Join(oc.GetCacheFilepath(), constants.OC_BINARY_NAME)
	cmdArgs := []string{"cluster", "up", "--use-existing-config"}

	// Set default value for host config, data and volumes
	viper.Set(hostConfigDir, viper.GetString(hostConfigDir))
	viper.Set(hostDataDir, viper.GetString(hostDataDir))
	viper.Set(hostVolumesDir, viper.GetString(hostVolumesDir))

	clusterUpFlagSet.VisitAll(func(flag *flag.Flag) {
		if viper.IsSet(flag.Name) {
			value := viper.GetString(flag.Name)
			key := flag.Name
			_, exists := minishiftToClusterUp[key]
			if exists {
				key = minishiftToClusterUp[key]
			}
			if !ocSupportFlag(cmdName, key) {
				glog.Errorf("Flag %s is not supported for oc version %s. Use 'openshift-version' flag to select a different version of OpenShift.", flag.Name, config.OpenShiftVersion)
				atexit.Exit(1)
			}
			cmdArgs = append(cmdArgs, "--"+key)
			cmdArgs = append(cmdArgs, value)
		}
	})

	err = runner.Run(cmdName, cmdArgs...)
	if err != nil {
		// TODO glog is probably not right here. Need some sort of logging wrapper
		glog.Errorln("Error starting the cluster: ", err)
		atexit.Exit(1)
	}
}

func validateOpenshiftVersion() {
	if viper.IsSet(openshiftVersion) {
		if !util.ValidateOpenshiftMinVersion(viper.GetString(openshiftVersion)) {
			fmt.Printf("Minishift does not support Openshift version %s. "+
				"You need to use a version >= 1.3.1.", viper.GetString(openshiftVersion))
			atexit.Exit(1)
		}
	}
}

func validateProxyArgs() {
	if viper.IsSet(httpProxy) {
		if !util.ValidateProxyURI(viper.GetString(httpProxy)) {
			glog.Exitf("HTTP Proxy URL is not valid, Please check help message")
		}
	}
	if viper.IsSet(httpsProxy) {
		if !util.ValidateProxyURI(viper.GetString(httpsProxy)) {
			glog.Exitf("HTTPS Proxy URL is not valid, Please check help message")
		}
	}
}

func ocSupportFlag(cmdName string, flag string) bool {
	cmdArgs := []string{"cluster", "up", "-h"}
	cmdOut, err := runner.Output(cmdName, cmdArgs...)
	if err != nil {
		glog.Errorf("Not able to get output of 'oc -h' Error: %s", err)
		atexit.Exit(1)
	}
	ocCommandOptions := minishiftUtil.ParseOcHelpCommand(cmdOut)
	if ocCommandOptions != nil {
		return minishiftUtil.FlagExist(ocCommandOptions, flag)
	}
	return false
}
