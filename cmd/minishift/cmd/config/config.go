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

package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/minishift/minishift/pkg/minikube/constants"
	validations "github.com/minishift/minishift/pkg/minishift/config"
)

type configFile interface {
	io.ReadWriter
}

type setFn func(string, string) error

type MinishiftConfig map[string]interface{}

type Setting struct {
	Name        string
	set         func(MinishiftConfig, string, string) error
	validations []setFn
	callbacks   []setFn
}

var settingsList []Setting

var (
	// minishift
	ISOUrl           = createConfigSetting("iso-url", SetString, []setFn{validations.IsValidUrl}, []setFn{RequiresRestartMsg}, true)
	CPUs             = createConfigSetting("cpus", SetInt, []setFn{validations.IsPositive}, []setFn{RequiresRestartMsg}, true)
	Memory           = createConfigSetting("memory", SetString, []setFn{validations.IsValidMemorySize}, []setFn{RequiresRestartMsg}, true)
	DiskSize         = createConfigSetting("disk-size", SetString, []setFn{validations.IsValidDiskSize}, []setFn{RequiresRestartMsg}, true)
	VmDriver         = createConfigSetting("vm-driver", SetString, []setFn{validations.IsValidDriver}, []setFn{RequiresRestartMsg}, true)
	OpenshiftVersion = createConfigSetting("openshift-version", SetString, nil, nil, true)
	HostOnlyCIDR     = createConfigSetting("host-only-cidr", SetString, []setFn{validations.IsValidCIDR}, nil, true)
	DockerEnv        = createConfigSetting("docker-env", SetSlice, nil, nil, true)
	DockerEngineOpt  = createConfigSetting("docker-opt", SetSlice, nil, nil, true)
	InsecureRegistry = createConfigSetting("insecure-registry", SetSlice, nil, nil, true)
	RegistryMirror   = createConfigSetting("registry-mirror", SetSlice, nil, nil, true)
	AddonEnv         = createConfigSetting("addon-env", SetSlice, nil, nil, true)

	// cluster up
	SkipRegistryCheck = createConfigSetting("skip-registry-check", SetBool, nil, nil, true)
	PublicHostname    = createConfigSetting("public-hostname", SetString, nil, nil, true)
	RoutingSuffix     = createConfigSetting("routing-suffix", SetString, nil, nil, true)
	HostConfigDir     = createConfigSetting("host-config-dir", SetString, []setFn{validations.IsValidPath}, nil, true)
	HostVolumeDir     = createConfigSetting("host-volumes-dir", SetString, []setFn{validations.IsValidPath}, nil, true)
	HostDataDir       = createConfigSetting("host-data-dir", SetString, []setFn{validations.IsValidPath}, nil, true)
	HostPvDir         = createConfigSetting("host-pv-dir", SetString, []setFn{validations.IsValidPath}, nil, true)
	ServerLogLevel    = createConfigSetting("server-loglevel", SetInt, []setFn{validations.IsPositive}, nil, true)
	OpenshiftEnv      = createConfigSetting("openshift-env", nil, nil, nil, false)
	Metrics           = createConfigSetting("metrics", SetBool, nil, nil, true)
	Logging           = createConfigSetting("logging", SetBool, nil, nil, true)
	// future enabled flags
	ServiceCatalog      = createConfigSetting("service-catalog", SetBool, nil, nil, true)
	ExtraClusterUpFlags = createConfigSetting("extra-clusterup-flags", SetString, nil, nil, true)

	// Setting proxy
	NoProxyList = createConfigSetting("no-proxy", SetString, nil, nil, true)
	HttpProxy   = createConfigSetting("http-proxy", SetString, []setFn{validations.IsValidProxy}, nil, true)
	HttpsProxy  = createConfigSetting("https-proxy", SetString, []setFn{validations.IsValidProxy}, nil, true)

	// Subscription Manager
	Username         = createConfigSetting("username", SetString, nil, nil, true)
	Password         = createConfigSetting("password", SetString, nil, nil, true)
	SkipRegistration = createConfigSetting("skip-registration", SetBool, nil, nil, true)

	// Global flags
	LogDir             = createConfigSetting("log_dir", SetString, []setFn{validations.IsValidPath}, nil, true)
	ShowLibmachineLogs = createConfigSetting("show-libmachine-logs", SetBool, nil, nil, true)

	// Host Folders
	HostFoldersMountPath = createConfigSetting("hostfolders-mountpath", SetString, nil, nil, true)
	HostFoldersAutoMount = createConfigSetting("hostfolders-automount", SetBool, nil, nil, true)

	// Image caching
	ImageCaching = createConfigSetting("image-caching", SetBool, nil, nil, true)
	CacheImages  = createConfigSetting("cache-images", SetSlice, nil, nil, false)

	// Pre-flight checks (before start)
	SkipCheckKVMDriver     = createConfigSetting("skip-check-kvm-driver", SetBool, nil, nil, true)
	WarnCheckKVMDriver     = createConfigSetting("warn-check-kvm-driver", SetBool, nil, nil, true)
	SkipCheckXHyveDriver   = createConfigSetting("skip-check-xhyve-driver", SetBool, nil, nil, true)
	WarnCheckXHyveDriver   = createConfigSetting("warn-check-xhyve-driver", SetBool, nil, nil, true)
	SkipCheckHyperVDriver  = createConfigSetting("skip-check-hyperv-driver", SetBool, nil, nil, true)
	WarnCheckHyperVDriver  = createConfigSetting("warn-check-hyperv-driver", SetBool, nil, nil, true)
	SkipCheckIsoUrl        = createConfigSetting("skip-check-iso-url", SetBool, nil, nil, true)
	WarnCheckIsoUrl        = createConfigSetting("warn-check-iso-url", SetBool, nil, nil, true)
	SkipCheckVMDriver      = createConfigSetting("skip-check-vm-driver", SetBool, nil, nil, true)
	WarnCheckVMDriver      = createConfigSetting("warn-check-vm-driver", SetBool, nil, nil, true)
	SkipCheckVBoxInstalled = createConfigSetting("skip-check-vbox-installed", SetBool, nil, nil, true)
	WarnCheckVBoxInstalled = createConfigSetting("warn-check-vbox-installed", SetBool, nil, nil, true)
	// Pre-flight checks (after start)
	SkipInstanceIP        = createConfigSetting("skip-check-instance-ip", SetBool, nil, nil, true)
	WarnInstanceIP        = createConfigSetting("warn-check-instance-ip", SetBool, nil, nil, true)
	SkipCheckNetworkHost  = createConfigSetting("skip-check-network-host", SetBool, nil, nil, true)
	WarnCheckNetworkHost  = createConfigSetting("warn-check-network-host", SetBool, nil, nil, true)
	SkipCheckNetworkPing  = createConfigSetting("skip-check-network-ping", SetBool, nil, nil, true)
	WarnCheckNetworkPing  = createConfigSetting("warn-check-network-ping", SetBool, nil, nil, true)
	SkipCheckNetworkHTTP  = createConfigSetting("skip-check-network-http", SetBool, nil, nil, true)
	WarnCheckNetworkHTTP  = createConfigSetting("warn-check-network-http", SetBool, nil, nil, true)
	SkipCheckStorageMount = createConfigSetting("skip-check-storage-mount", SetBool, nil, nil, true)
	WarnCheckStorageMount = createConfigSetting("warn-check-storage-mount", SetBool, nil, nil, true)
	SkipCheckStorageUsage = createConfigSetting("skip-check-storage-usage", SetBool, nil, nil, true)
	WarnCheckStorageUsage = createConfigSetting("warn-check-storage-usage", SetBool, nil, nil, true)

	// Pre-flight values
	CheckNetworkHttpHost = createConfigSetting("check-network-http-host", SetString, nil, nil, true)
	CheckNetworkPingHost = createConfigSetting("check-network-ping-host", SetString, nil, nil, true)

	// Network settings (Hyper-V only)
	NetworkDevice = createConfigSetting("network-device", SetString, nil, nil, true)
	IPAddress     = createConfigSetting("network-ipaddress", SetString, []setFn{validations.IsValidIPv4Address}, nil, true)
	Netmask       = createConfigSetting("network-netmask", SetString, []setFn{validations.IsValidNetmask}, nil, true)
	Gateway       = createConfigSetting("network-gateway", SetString, []setFn{validations.IsValidIPv4Address}, nil, true)
	NameServer    = createConfigSetting("network-nameserver", SetString, []setFn{validations.IsValidIPv4Address}, nil, true)
)

func createConfigSetting(name string, set func(MinishiftConfig, string, string) error, validations []setFn, callbacks []setFn, isApply bool) *Setting {
	flag := Setting{
		Name:        name,
		set:         set,
		validations: validations,
		callbacks:   callbacks,
	}
	if isApply {
		settingsList = append(settingsList, flag)
	}
	return &flag
}

var ConfigCmd = &cobra.Command{
	Use:   "config SUBCOMMAND [flags]",
	Short: "Modifies Minishift configuration properties.",
	Long: `Modifies Minishift configuration properties. Some of the configuration properties are equivalent
to the options that you set when you run the 'minishift start' command.

Configurable properties (enter as SUBCOMMAND): ` + "\n\n" + configurableFields(),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func configurableFields() string {
	var fields []string
	for _, s := range settingsList {
		fields = append(fields, " * "+s.Name)
	}
	return strings.Join(fields, "\n")
}

// ReadConfig reads the config from $MINISHIFT_HOME/config/config.json file
func ReadConfig() (MinishiftConfig, error) {
	f, err := os.Open(constants.ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("Cannot open file '%s': %s", constants.ConfigFile, err)
	}
	var m MinishiftConfig
	m, err = decode(f)
	if err != nil {
		return nil, fmt.Errorf("Cannot decode config '%s': %s", constants.ConfigFile, err)
	}

	return m, nil
}

// Writes a config to the $MINISHIFT_HOME/config/config.json file
func WriteConfig(m MinishiftConfig) error {
	f, err := os.Create(constants.ConfigFile)
	if err != nil {
		return fmt.Errorf("Cannot create file '%s': %s", constants.ConfigFile, err)
	}
	defer f.Close()
	err = encode(f, m)
	if err != nil {
		return fmt.Errorf("Cannot encode config '%s': %s", constants.ConfigFile, err)
	}
	return nil
}

func decode(r io.Reader) (MinishiftConfig, error) {
	var data MinishiftConfig
	err := json.NewDecoder(r).Decode(&data)
	return data, err
}

func encode(w io.Writer, m MinishiftConfig) error {
	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}

	_, err = w.Write(b)

	return err
}
