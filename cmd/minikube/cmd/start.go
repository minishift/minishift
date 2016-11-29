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
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift"
	"github.com/minishift/minishift/pkg/util"
	dockerhost "github.com/openshift/origin/pkg/bootstrap/docker/host"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	isoURL                = "iso-url"
	memory                = "memory"
	cpus                  = "cpus"
	humanReadableDiskSize = "disk-size"
	vmDriver              = "vm-driver"
	openshiftVersion      = "openshift-version"
	hostOnlyCIDR          = "host-only-cidr"
	deployRegistry        = "deploy-registry"
	deployRouter          = "deploy-router"
)

var (
	dockerEnv        []string
	insecureRegistry []string
	registryMirror   []string
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a local OpenShift cluster.",
	Long: `Starts a local OpenShift cluster using Virtualbox. This command
assumes you already have Virtualbox installed.`,
	Run: runStart,
}

// ClientStartConfig is the configuration for the client start command
type ClientStartConfig struct {
	ImageVersion              string
	Image                     string
	DockerMachine             string
	ShouldCreateDockerMachine bool
	SkipRegistryCheck         bool
	ShouldInstallMetrics      bool
	PortForwarding            bool

	UseNsenterMount    bool
	SetPropagationMode bool
	HostName           string
	ServerIP           string
	CACert             string
	PublicHostname     string
	RoutingSuffix      string
	DNSPort            int

	LocalConfigDir    string
	HostVolumesDir    string
	HostConfigDir     string
	HostDataDir       string
	UseExistingConfig bool
	Environment       []string
	ServerLogLevel    int

	usingDefaultImages         bool
	usingDefaultOpenShiftImage bool
}

var clusterUpConfig = ClientStartConfig{}
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
	fmt.Println("Starting local OpenShift cluster...")

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
		DeployRouter:     viper.GetBool(deployRouter),
		DeployRegistry:   viper.GetBool(deployRegistry),
		OpenShiftVersion: viper.GetString(openshiftVersion),
	}

	var host *host.Host
	start := func() (err error) {
		host, err = cluster.StartHost(libMachineClient, config)
		if err != nil {
			glog.Errorf("Error starting host: %s. Retrying.\n", err)
		}
		return err
	}
	err := util.Retry(3, start)
	if err != nil {
		glog.Errorln("Error starting host: ", err)
		os.Exit(1)
	}

	// Making sure the required Docker environment variables are set to make 'cluster up' work
	envMap, err := cluster.GetHostDockerEnv(libMachineClient)
	for k, v := range envMap {
		os.Setenv(k, v)
	}

	clusterUp()
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
	startCmd.Flags().String(isoURL, constants.DefaultIsoUrl, "Location of the minishift iso")
	startCmd.Flags().String(vmDriver, constants.DefaultVMDriver, fmt.Sprintf("VM driver is one of: %v", constants.SupportedVMDrivers))
	startCmd.Flags().Int(memory, constants.DefaultMemory, "Amount of RAM allocated to the minishift VM")
	startCmd.Flags().Int(cpus, constants.DefaultCPUS, "Number of CPUs allocated to the minishift VM")
	startCmd.Flags().String(humanReadableDiskSize, constants.DefaultDiskSize, "Disk size allocated to the minishift VM (format: <number>[<unit>], where unit = b, k, m or g)")
	startCmd.Flags().String(hostOnlyCIDR, "192.168.99.1/24", "The CIDR to be used for the minishift VM (only supported with Virtualbox driver)")
	startCmd.Flags().StringSliceVar(&dockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. (format: key=value)")
	startCmd.Flags().StringSliceVar(&insecureRegistry, "insecure-registry", []string{"172.30.0.0/16"}, "Insecure Docker registries to pass to the Docker daemon")
	startCmd.Flags().StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon")
	startCmd.Flags().Bool(deployRegistry, true, "Should the OpenShift internal Docker registry be deployed?")
	startCmd.Flags().Bool(deployRouter, false, "Should the OpenShift router be deployed?")
	startCmd.Flags().String(openshiftVersion, "", "The OpenShift version that the minishift VM will run (ex: v1.2.3) OR a URI which contains an openshift binary (ex: file:///home/developer/go/src/github.com/openshift/origin/_output/local/bin/linux/amd64/openshift)")

	// TODO Determine which flags we need to expose via minishift and which flags we can hard-wire (HF)
	startCmd.Flags().BoolVar(&clusterUpConfig.ShouldCreateDockerMachine, "create-machine", false, "Create a Docker machine if one doesn't exist")
	startCmd.Flags().StringVar(&clusterUpConfig.DockerMachine, "docker-machine", "", "Specify the Docker machine to use")
	startCmd.Flags().StringVar(&clusterUpConfig.ImageVersion, "version", "", "Specify the tag for OpenShift images")
	startCmd.Flags().StringVar(&clusterUpConfig.Image, "image", "openshift/origin", "Specify the images to use for OpenShift")
	startCmd.Flags().BoolVar(&clusterUpConfig.SkipRegistryCheck, "skip-registry-check", false, "Skip Docker daemon registry check")
	startCmd.Flags().StringVar(&clusterUpConfig.PublicHostname, "public-hostname", "", "Public hostname for OpenShift cluster")
	startCmd.Flags().StringVar(&clusterUpConfig.RoutingSuffix, "routing-suffix", "", "Default suffix for server routes")
	startCmd.Flags().BoolVar(&clusterUpConfig.UseExistingConfig, "use-existing-config", false, "Use existing configuration if present")
	startCmd.Flags().StringVar(&clusterUpConfig.HostConfigDir, "host-config-dir", dockerhost.DefaultConfigDir, "Directory on Docker host for OpenShift configuration")
	startCmd.Flags().StringVar(&clusterUpConfig.HostVolumesDir, "host-volumes-dir", dockerhost.DefaultVolumesDir, "Directory on Docker host for OpenShift volumes")
	startCmd.Flags().StringVar(&clusterUpConfig.HostDataDir, "host-data-dir", "", "Directory on Docker host for OpenShift data. If not specified, etcd data will not be persisted on the host.")
	startCmd.Flags().BoolVar(&clusterUpConfig.PortForwarding, "forward-ports", clusterUpConfig.PortForwarding, "Use Docker port-forwarding to communicate with origin container. Requires 'socat' locally.")
	startCmd.Flags().IntVar(&clusterUpConfig.ServerLogLevel, "server-loglevel", 0, "Log level for OpenShift server")
	startCmd.Flags().StringSliceVarP(&clusterUpConfig.Environment, "env", "e", clusterUpConfig.Environment, "Specify key value pairs of environment variables to set on OpenShift container")
	startCmd.Flags().BoolVar(&clusterUpConfig.ShouldInstallMetrics, "metrics", false, "Install metrics (experimental)")

	viper.BindPFlags(startCmd.Flags())

	RootCmd.AddCommand(startCmd)
}

// clusterUp downloads and installs the oc binary in order to run 'cluster up'
func clusterUp() {
	// TODO find latest stable release programatically
	oc := minishift.Oc{"v1.3.1", filepath.Join(constants.Minipath, "cache")}
	err := oc.EnsureIsCached()
	if err != nil {
		glog.Errorln("Error starting 'cluster up': ", err)
		os.Exit(1)
	}

	cmdName := filepath.Join(oc.GetCacheFilepath(), minishift.OC_BINARY_NAME)
	// TODO pass along relevant options
	cmdArgs := []string{"cluster", "up"}

	err = runner.Run(cmdName, cmdArgs...)
	if err != nil {
		// TODO glog is probably not right here. Need some sort of logging wrapper
		glog.Errorln("Error starting 'cluster up': ", err)
		os.Exit(1)
	}
}
