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

	"github.com/minishift/minishift/pkg/minikube/constants"
	validations "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DefaultRoutingSuffix = ".nip.io"
)

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
	ISOUrl                = createConfigSetting("iso-url", SetString, []setFn{validations.IsValidISOUrl}, []setFn{RequiresRestartMsg}, true, nil)
	CPUs                  = createConfigSetting("cpus", SetInt, []setFn{validations.IsPositive}, []setFn{RequiresRestartMsg}, true, nil)
	Memory                = createConfigSetting("memory", SetString, []setFn{validations.IsValidMemorySize}, []setFn{RequiresRestartMsg}, true, nil)
	DiskSize              = createConfigSetting("disk-size", SetString, []setFn{validations.IsValidDiskSize}, []setFn{RequiresRestartMsg}, true, nil)
	VmDriver              = createConfigSetting("vm-driver", SetString, []setFn{validations.IsValidDriver}, []setFn{RequiresRestartMsg}, true, nil)
	OpenshiftVersion      = createConfigSetting("openshift-version", SetString, nil, nil, true, nil)
	HostOnlyCIDR          = createConfigSetting("host-only-cidr", SetString, []setFn{validations.IsValidCIDR}, nil, true, nil)
	DockerEnv             = createConfigSetting("docker-env", SetSlice, nil, nil, true, nil)
	DockerEngineOpt       = createConfigSetting("docker-opt", SetSlice, nil, nil, true, nil)
	InsecureRegistry      = createConfigSetting("insecure-registry", SetSlice, nil, nil, true, nil)
	RegistryMirror        = createConfigSetting("registry-mirror", SetSlice, nil, nil, true, nil)
	AddonEnv              = createConfigSetting("addon-env", SetSlice, nil, nil, true, nil)
	RemoteIPAddress       = createConfigSetting("remote-ipaddress", SetString, nil, nil, true, nil)
	RemoteSSHUser         = createConfigSetting("remote-ssh-user", SetString, nil, nil, true, nil)
	SSHKeyToConnectRemote = createConfigSetting("remote-ssh-key", SetString, nil, nil, true, nil)

	// cluster up
	SkipRegistryCheck = createConfigSetting("skip-registry-check", SetBool, nil, nil, true, nil)
	PublicHostname    = createConfigSetting("public-hostname", SetString, nil, nil, true, nil)
	RoutingSuffix     = createConfigSetting("routing-suffix", SetString, nil, nil, true, nil)
	ServerLogLevel    = createConfigSetting("server-loglevel", SetInt, []setFn{validations.IsPositive}, nil, true, nil)

	// future enabled flags
	ExtraClusterUpFlags = createConfigSetting("extra-clusterup-flags", SetString, nil, nil, true, nil)

	// Setting proxy
	NoProxyList = createConfigSetting("no-proxy", SetString, nil, nil, true, nil)
	HttpProxy   = createConfigSetting("http-proxy", SetString, []setFn{validations.IsValidProxy}, nil, true, nil)
	HttpsProxy  = createConfigSetting("https-proxy", SetString, []setFn{validations.IsValidProxy}, nil, true, nil)

	// Subscription Manager
	Username         = createConfigSetting("username", SetString, nil, nil, true, nil)
	Password         = createConfigSetting("password", SetString, nil, nil, true, nil)
	SkipRegistration = createConfigSetting("skip-registration", SetBool, nil, nil, true, nil)

	// Global flags
	LogDir             = createConfigSetting("log_dir", SetString, []setFn{validations.IsValidPath}, nil, true, nil)
	ShowLibmachineLogs = createConfigSetting("show-libmachine-logs", SetBool, nil, nil, true, nil)

	// Host Folders
	HostFoldersMountPath = createConfigSetting("hostfolders-mountpath", SetString, nil, nil, true, nil)
	HostFoldersAutoMount = createConfigSetting("hostfolders-automount", SetBool, nil, nil, true, nil)
	HostFoldersSftpPort  = createConfigSetting("hostfolders-sftp-port", SetInt, []setFn{validations.IsValidPort}, nil, true, nil)

	// No Provision
	NoProvision = createConfigSetting("no-provision", SetBool, nil, nil, true, nil)

	// Image caching
	ImageCaching = createConfigSetting("image-caching", SetBool, nil, nil, true, true)
	CacheImages  = createConfigSetting("cache-images", SetSlice, nil, nil, false, nil)

	// Pre-flight checks (before start)
	SkipDeprecationCheck      = createConfigSetting("skip-check-deprecation", SetBool, nil, nil, true, nil)
	WarnDeprecationCheck      = createConfigSetting("warn-check-deprecation", SetBool, nil, nil, true, true)
	SkipCheckKVMDriver        = createConfigSetting("skip-check-kvm-driver", SetBool, nil, nil, true, nil)
	WarnCheckKVMDriver        = createConfigSetting("warn-check-kvm-driver", SetBool, nil, nil, true, false)
	SkipCheckXHyveDriver      = createConfigSetting("skip-check-xhyve-driver", SetBool, nil, nil, true, nil)
	WarnCheckXHyveDriver      = createConfigSetting("warn-check-xhyve-driver", SetBool, nil, nil, true, false)
	SkipCheckHyperVDriver     = createConfigSetting("skip-check-hyperv-driver", SetBool, nil, nil, true, nil)
	WarnCheckHyperVDriver     = createConfigSetting("warn-check-hyperv-driver", SetBool, nil, nil, true, false)
	SkipCheckIsoUrl           = createConfigSetting("skip-check-iso-url", SetBool, nil, nil, true, nil)
	WarnCheckIsoUrl           = createConfigSetting("warn-check-iso-url", SetBool, nil, nil, true, false)
	SkipCheckVMDriver         = createConfigSetting("skip-check-vm-driver", SetBool, nil, nil, true, nil)
	WarnCheckVMDriver         = createConfigSetting("warn-check-vm-driver", SetBool, nil, nil, true, false)
	SkipCheckVBoxInstalled    = createConfigSetting("skip-check-vbox-installed", SetBool, nil, nil, true, nil)
	WarnCheckVBoxInstalled    = createConfigSetting("warn-check-vbox-installed", SetBool, nil, nil, true, false)
	SkipCheckOpenShiftVersion = createConfigSetting("skip-check-openshift-version", SetBool, nil, nil, true, nil)
	WarnCheckOpenShiftVersion = createConfigSetting("warn-check-openshift-version", SetBool, nil, nil, true, false)
	SkipCheckOpenShiftRelease = createConfigSetting("skip-check-openshift-release", SetBool, nil, nil, true, nil)
	WarnCheckOpenShiftRelease = createConfigSetting("warn-check-openshift-release", SetBool, nil, nil, true, false)
	SkipPreflightChecks       = createConfigSetting("skip-startup-checks", SetBool, nil, nil, true, false)

	// Pre-flight checks for artifacts (before start)
	SkipCheckClusterUpFlag = createConfigSetting("skip-check-clusterup-flags", SetBool, nil, nil, true, nil)
	WarnCheckClusterUpFlag = createConfigSetting("warn-check-clusterup-flags", SetBool, nil, nil, true, nil)

	// Pre-flight checks (after start)
	SkipInstanceIP        = createConfigSetting("skip-check-instance-ip", SetBool, nil, nil, true, nil)
	WarnInstanceIP        = createConfigSetting("warn-check-instance-ip", SetBool, nil, nil, true, false)
	SkipCheckNetworkHost  = createConfigSetting("skip-check-network-host", SetBool, nil, nil, true, nil)
	WarnCheckNetworkHost  = createConfigSetting("warn-check-network-host", SetBool, nil, nil, true, false)
	SkipCheckNetworkPing  = createConfigSetting("skip-check-network-ping", SetBool, nil, nil, true, nil)
	WarnCheckNetworkPing  = createConfigSetting("warn-check-network-ping", SetBool, nil, nil, true, true)
	SkipCheckNetworkHTTP  = createConfigSetting("skip-check-network-http", SetBool, nil, nil, true, nil)
	WarnCheckNetworkHTTP  = createConfigSetting("warn-check-network-http", SetBool, nil, nil, true, true)
	SkipCheckStorageMount = createConfigSetting("skip-check-storage-mount", SetBool, nil, nil, true, nil)
	WarnCheckStorageMount = createConfigSetting("warn-check-storage-mount", SetBool, nil, nil, true, false)
	SkipCheckStorageUsage = createConfigSetting("skip-check-storage-usage", SetBool, nil, nil, true, nil)
	WarnCheckStorageUsage = createConfigSetting("warn-check-storage-usage", SetBool, nil, nil, true, false)
	SkipCheckNameservers  = createConfigSetting("skip-check-nameservers", SetBool, nil, nil, true, nil)
	WarnCheckNameservers  = createConfigSetting("warn-check-nameservers", SetBool, nil, nil, true, false)

	// Pre-flight values
	CheckNetworkHttpHost = createConfigSetting("check-network-http-host", SetString, nil, nil, true, "http://minishift.io/index.html")
	CheckNetworkPingHost = createConfigSetting("check-network-ping-host", SetString, nil, nil, true, "8.8.8.8")

	// Network settings (Hyper-V only)
	NetworkDevice = createConfigSetting("network-device", SetString, nil, nil, true, nil)
	IPAddress     = createConfigSetting("network-ipaddress", SetString, []setFn{validations.IsValidIPv4Address}, nil, true, nil)
	Netmask       = createConfigSetting("network-netmask", SetString, []setFn{validations.IsValidNetmask}, nil, true, nil)
	Gateway       = createConfigSetting("network-gateway", SetString, []setFn{validations.IsValidIPv4Address}, nil, true, nil)

	// Network setting
	NameServers           = createConfigSetting("network-nameserver", SetSlice, []setFn{validations.IsValidIPv4AddressSlice}, nil, true, nil)
	DnsmasqContainerized  = createConfigSetting("network-dnsmasq-containerized", SetBool, nil, nil, true, false)
	DnsmasqContainerImage = createConfigSetting("network-dnsmasq-container", SetString, nil, nil, true, nil)

	// Hyper-V
	HypervVirtualSwitch = createConfigSetting("hyperv-virtual-switch", SetString, []setFn{validations.IsValidHypervVirtualSwitch}, nil, true, nil)
)

func createConfigSetting(name string, set func(MinishiftConfig, string, string) error, validations []setFn, callbacks []setFn, isApply bool, defaultVal interface{}) *Setting {
	flag := Setting{
		Name:        name,
		set:         set,
		validations: validations,
		callbacks:   callbacks,
	}
	if isApply {
		settingsList = append(settingsList, flag)
		if defaultVal != nil {
			viper.SetDefault(name, defaultVal)
		}
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
