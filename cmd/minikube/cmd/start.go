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
	Long: `Starts a local OpenShift cluster using VirtualBox. You must have VirtualBox installed
	to start the cluster. If you have an existing configuration file, you can use it by setting the
	--use-existing-config option.`,
	//NEEDINFO: There are lots of options, we should give examples for different use-cases. Also, does this create a new cluster or just spin up an existing cluster?
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
	fmt.Println("Starting the local OpenShift cluster...")

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
		host, err = cluster.StartHost(libMachineClient, config)
		if err != nil {
			glog.Errorf("Error starting the host: %s. Retrying.\n", err)
		}
		return err
	}
	err := util.Retry(3, start)
	if err != nil {
		glog.Errorln("Error starting the host: ", err)
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

// initClusterUpFlags creates the CLI flags which needs to be passed on to 'oc cluster up'
func initStartFlags() {
	startFlagSet.String(isoURL, constants.DefaultIsoUrl, "Location of the minishift ISO.")
	//NEEDINFO: Is this the same as the binary? we should be consistent in how we refer to it. Also, we need to change the default from jdyson.
	startFlagSet.String(vmDriver, constants.DefaultVMDriver, fmt.Sprintf("The driver to use for the Minishift VM. Possible values: %v", constants.SupportedVMDrivers))
	//NEEDINFO: If virtualbox is the virt provider how come kvm is the default driver and how are drivers related to this at all?
	startFlagSet.Int(memory, constants.DefaultMemory, "Amount of RAM to allocate to the Minishift VM.")
	startFlagSet.Int(cpus, constants.DefaultCPUS, "Number of CPU cores to allocate to the Minishift VM.")
	startFlagSet.String(humanReadableDiskSize, constants.DefaultDiskSize, "Disk size to allocate to the Minishift VM. Use the format <size><unit>, where unit = b, k, m or g.")
	startFlagSet.String(hostOnlyCIDR, "192.168.99.1/24", "The CIDR to be used for the minishift VM (only supported with Virtualbox driver)")
	//NEEDINFO: What is a CIDR?
	startFlagSet.StringSliceVar(&dockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. Use the format <key>=<value>.")
	//NEEDINFO: How do you separate multiple env vars? Comma? Space + comma? etc..
	startFlagSet.StringSliceVar(&insecureRegistry, "insecure-registry", []string{"172.30.0.0/16"}, "Insecure Docker registries to pass to the Docker daemon")
	//NEEDINFO: "insecure" is a really inappropriate word, not sure what it's supposed to mean but we really cannot use it.
	startFlagSet.StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon.")
	//NEEDINFO: Is this URLs of the mirrors? hostnames? also, there's a blank default and that looks weird, either remove it or populate it
	startFlagSet.String(openshiftVersion, version.GetOpenShiftVersion(), "The OpenShift version to run. Use the format v<n.n.n>")
}

// initClusterUpFlags creates the CLI flags which needs to be passed on to 'oc cluster up'
func initClusterUpFlags() {
	//clusterUpFlagSet.StringVar(&clusterUpConfig.Image, "image", "openshift/origin", "Specify the images to use for OpenShift")
	clusterUpFlagSet.Bool(skipRegistryCheck, false, "Skip the Docker daemon registry check.")
	clusterUpFlagSet.String(publicHostname, "", "Public host name of the OpenShift cluster.")
	clusterUpFlagSet.String(routingSuffix, "", "Default suffix for the server routes.")
	clusterUpFlagSet.Bool(useExistingConfig, false, "Use existing configuration if available.")
	clusterUpFlagSet.String(hostConfigDir, dockerhost.DefaultConfigDir, "Location of the OpenShift configuration files on the host.")
	clusterUpFlagSet.String(hostVolumesDir, dockerhost.DefaultVolumesDir, "Location of the OpenShift volumes on the host.")
	clusterUpFlagSet.String(hostDataDir, "", "Location of the OpenShift data on the host. If not specified, etcd data will not be persisted on the host.")
	//NEEDINFO: what does etcd data mean? persistent?
	clusterUpFlagSet.Bool(forwardPorts, false, "Use Docker port forwarding to communicate with origin container. Requires 'socat' locally.")
	//NEEDINFO: origin or original? origin might not be the right word here. what is socat?
	clusterUpFlagSet.Int(serverLogLevel, 0, "Set the log level for the OpenShift server.")
	//NEEDINFO: do we need to default and possible values here? what is this used for, debugging openshift? the app? minishift?
	clusterUpFlagSet.StringSliceVarP(&openShiftEnv, openshiftEnv, "e", []string{}, "Specify key-value pairs of environment variables to pass to the OpenShift container.")
	//NEEDINFO: This is the first time I see OpenShift container, isn't this the same as Minishift VM? or app? and what's the difference between this and the Docker daemon env vars?
	clusterUpFlagSet.Bool(metrics, false, "Install metrics (experimental)")
	//NEEDINFO: What is this? do we even support this?
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

	cmdName := filepath.Join(oc.GetCacheFilepath(), cache.OC_BINARY_NAME)
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
		glog.Errorln("Error starting the cluster: ", err)
		os.Exit(1)
	}
}
