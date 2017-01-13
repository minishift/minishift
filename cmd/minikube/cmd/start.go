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
	"github.com/minishift/minishift/pkg/minishift/provisioner"
	"github.com/minishift/minishift/pkg/minishift/registration"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/version"
	dockerhost "github.com/openshift/origin/pkg/bootstrap/docker/host"
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
)

var (
	dockerEnv        []string
	insecureRegistry []string
	registryMirror   []string
	openShiftEnv     []string
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
	hostConfigDirectory = "/var/lib/minishift/openshift.local.config"
	hostDataDirectory   = "/var/lib/minishift/hostdata"
)

// startFlagSet contains the minishift specific command line switches
var startFlagSet = flag.NewFlagSet(commandName, flag.ContinueOnError)

// clusterUpFlagSet contains the command line switches which needs to be passed on to 'cluster up'
var clusterUpFlagSet = flag.NewFlagSet(commandName, flag.ContinueOnError)

// minishiftToClusterUp is a mapping between falg names used in minishift CLI and flag name as passed to 'cluster up'
var minishiftToClusterUp = map[string]string{
	"openshift-env": "env",
}

// runner executes commands on the host
var runner util.Runner = &util.RealRunner{}

func SetRunner(newRunner util.Runner) {
	runner = newRunner
}

// TODO Figure out what I cannot use constants.Minipath in the test - http://stackoverflow.com/questions/37284423/glog-flag-redefined-error
func SetMinishiftDir(newDir string) {
	constants.Minipath = newDir
}

// runStart is executed as part of the start command
func runStart(cmd *cobra.Command, args []string) {
	libMachineClient := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer libMachineClient.Close()

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

	fmt.Printf("Starting the local OpenShift cluster using '%s' hypervisor...\n", config.VMDriver)

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
		os.Exit(1)
	}
	// Register Host VM
	if err := registration.RegisterHostVM(host, RegistrationParameters); err != nil {
		fmt.Printf("Error registering the VM: %s", err)
		os.Exit(1)
	}

	// Making sure the required Docker environment variables are set to make 'cluster up' work
	envMap, err := cluster.GetHostDockerEnv(libMachineClient)
	for k, v := range envMap {
		os.Setenv(k, v)
	}

	clusterUp(&config)
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

	startCmd.Flags().AddFlagSet(startFlagSet)
	startCmd.Flags().AddFlagSet(clusterUpFlagSet)

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
	startFlagSet.StringSliceVar(&dockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. Use the format <key>=<value>.")
	startFlagSet.StringSliceVar(&insecureRegistry, "insecure-registry", []string{"172.30.0.0/16"}, "Non-secure Docker registries to pass to the Docker daemon.")
	startFlagSet.StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon.")
	startFlagSet.String(openshiftVersion, version.GetOpenShiftVersion(), "The OpenShift version to run, eg. v1.3.1")
}

// initClusterUpFlags creates the CLI flags which needs to be passed on to 'oc cluster up'
func initClusterUpFlags() {
	//clusterUpFlagSet.StringVar(&clusterUpConfig.Image, "image", "openshift/origin", "Specify the images to use for OpenShift")
	clusterUpFlagSet.Bool(skipRegistryCheck, false, "Skip the Docker daemon registry check.")
	clusterUpFlagSet.String(publicHostname, "", "Public hostname of the OpenShift cluster.")
	clusterUpFlagSet.String(routingSuffix, "", "Default suffix for the server routes.")
	clusterUpFlagSet.String(hostConfigDir, hostConfigDirectory, "Location of the OpenShift configuration on the Docker host.")
	clusterUpFlagSet.String(hostVolumesDir, dockerhost.DefaultVolumesDir, "Location of the OpenShift volumes on the Docker host.")
	clusterUpFlagSet.String(hostDataDir, hostDataDirectory, "Location of the OpenShift data on the Docker host. If not specified, etcd data will not be persisted on the host.")
	clusterUpFlagSet.Bool(forwardPorts, false, "Use Docker port forwarding to communicate with the origin container. Requires 'socat' locally.")
	clusterUpFlagSet.Int(serverLogLevel, 0, "Log level for the OpenShift server.")
	clusterUpFlagSet.StringSliceVarP(&openShiftEnv, openshiftEnv, "e", []string{}, "Specify key-value pairs of environment variables to set on the OpenShift container.")
	clusterUpFlagSet.Bool(metrics, false, "Install metrics (experimental)")
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
		os.Exit(1)
	}

	cmdName := filepath.Join(oc.GetCacheFilepath(), constants.OC_BINARY_NAME)
	cmdArgs := []string{"cluster", "up", "--use-existing-config"}

	// Set default value for host config and data
	viper.Set(hostConfigDir, viper.GetString(hostConfigDir))
	viper.Set(hostDataDir, viper.GetString(hostDataDir))

	clusterUpFlagSet.VisitAll(func(flag *flag.Flag) {
		if viper.IsSet(flag.Name) {
			value := viper.GetString(flag.Name)
			key := flag.Name
			_, exists := minishiftToClusterUp[key]
			if exists {
				key = minishiftToClusterUp[key]
			}
			cmdArgs = append(cmdArgs, "--"+key)
			cmdArgs = append(cmdArgs, value)
		}
	})

	err = runner.Run(cmdName, cmdArgs...)
	if err != nil {
		// TODO glog is probably not right here. Need some sort of logging wrapper
		glog.Errorln("Error starting the cluster: ", err)
		os.Exit(1)
	}
}
