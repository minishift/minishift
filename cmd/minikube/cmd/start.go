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
	useExistingConfig = "use-existing-config"
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
	Long: `Starts a local OpenShift cluster using Virtualbox. This command
assumes you already have Virtualbox installed.`,
	Run: runStart,
}

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

	var host *host.Host
	start := func() (err error) {
		fmt.Printf("Starting local OpenShift instance using '%s' hypervisor...\n", config.VMDriver)
		host, err = cluster.StartHost(libMachineClient, config)
		if err != nil {
			glog.Errorf("Error starting machine: %s. Retrying.\n", err)
		}
		return err
	}
	err := util.Retry(3, start)
	if err != nil {
		glog.Errorln("Error starting machine: ", err)
		os.Exit(1)
	}
	// Register Host VM
	if err := registration.RegisterHostVM(host, RegistrationParameters); err != nil {
		fmt.Printf("Error registering machine: %s", err)
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
		glog.Errorf("Invalid disk size: %s", err)
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

// initClusterUpFlags creates the CLI flags which needs to be passed on to 'oc cluster up'
func initStartFlags() {
	startFlagSet.String(isoURL, constants.DefaultIsoUrl, "Location of the minishift iso")
	startFlagSet.String(vmDriver, constants.DefaultVMDriver, fmt.Sprintf("VM driver is one of: %v", constants.SupportedVMDrivers))
	startFlagSet.Int(memory, constants.DefaultMemory, "Amount of RAM allocated to the minishift VM")
	startFlagSet.Int(cpus, constants.DefaultCPUS, "Number of CPUs allocated to the minishift VM")
	startFlagSet.String(humanReadableDiskSize, constants.DefaultDiskSize, "Disk size allocated to the minishift VM (format: <number>[<unit>], where unit = b, k, m or g)")
	startFlagSet.String(hostOnlyCIDR, "192.168.99.1/24", "The CIDR to be used for the minishift VM (only supported with Virtualbox driver)")
	startFlagSet.StringSliceVar(&dockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. (format: key=value)")
	startFlagSet.StringSliceVar(&insecureRegistry, "insecure-registry", []string{"172.30.0.0/16"}, "Insecure Docker registries to pass to the Docker daemon")
	startFlagSet.StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon")
	startFlagSet.String(openshiftVersion, version.GetOpenShiftVersion(), "The OpenShift version that the minishift VM will run (ex: v1.2.3)")
}

// initClusterUpFlags creates the CLI flags which needs to be passed on to 'oc cluster up'
func initClusterUpFlags() {
	//clusterUpFlagSet.StringVar(&clusterUpConfig.Image, "image", "openshift/origin", "Specify the images to use for OpenShift")
	clusterUpFlagSet.Bool(skipRegistryCheck, false, "Skip Docker daemon registry check")
	clusterUpFlagSet.String(publicHostname, "", "Public hostname for OpenShift cluster")
	clusterUpFlagSet.String(routingSuffix, "", "Default suffix for server routes")
	clusterUpFlagSet.Bool(useExistingConfig, false, "Use existing configuration if present")
	clusterUpFlagSet.String(hostConfigDir, dockerhost.DefaultConfigDir, "Directory on Docker host for OpenShift configuration")
	clusterUpFlagSet.String(hostVolumesDir, dockerhost.DefaultVolumesDir, "Directory on Docker host for OpenShift volumes")
	clusterUpFlagSet.String(hostDataDir, "", "Directory on Docker host for OpenShift data. If not specified, etcd data will not be persisted on the host.")
	clusterUpFlagSet.Bool(forwardPorts, false, "Use Docker port-forwarding to communicate with origin container. Requires 'socat' locally.")
	clusterUpFlagSet.Int(serverLogLevel, 0, "Log level for OpenShift server")
	clusterUpFlagSet.StringSliceVarP(&openShiftEnv, openshiftEnv, "e", []string{}, "Specify key value pairs of environment variables to set on OpenShift container")
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
		glog.Errorln("Error starting 'cluster up': ", err)
		os.Exit(1)
	}

	cmdName := filepath.Join(oc.GetCacheFilepath(), constants.OC_BINARY_NAME)
	cmdArgs := []string{"cluster", "up"}
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
		glog.Errorln("Error starting 'cluster up': ", err)
		os.Exit(1)
	}
}
