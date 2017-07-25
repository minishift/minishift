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

const (
	B2dIsoAlias    = "b2d"
	CentOsIsoAlias = "centos"
)

var (
	// minishift
	ISOUrl           = createFlag("iso-url", SetString, []setFn{IsValidUrl}, []setFn{RequiresRestartMsg}, true)
	CPUs             = createFlag("cpus", SetInt, []setFn{IsPositive}, []setFn{RequiresRestartMsg}, true)
	Memory           = createFlag("memory", SetInt, []setFn{IsPositive}, []setFn{RequiresRestartMsg}, true)
	DiskSize         = createFlag("disk-size", SetString, []setFn{IsValidDiskSize}, []setFn{RequiresRestartMsg}, true)
	VmDriver         = createFlag("vm-driver", SetString, []setFn{IsValidDriver}, []setFn{RequiresRestartMsg}, true)
	OpenshiftVersion = createFlag("openshift-version", SetString, nil, nil, true)
	HostOnlyCIDR     = createFlag("host-only-cidr", SetString, []setFn{IsValidCIDR}, nil, true)
	DockerEnv        = createFlag("docker-env", SetSlice, nil, nil, true)
	DockerEngineOpt  = createFlag("docker-opt", SetSlice, nil, nil, true)
	InsecureRegistry = createFlag("insecure-registry", SetSlice, nil, nil, true)
	RegistryMirror   = createFlag("registry-mirror", SetSlice, nil, nil, true)
	AddonEnv         = createFlag("addon-env", SetSlice, nil, nil, true)

	// cluster up
	SkipRegistryCheck = createFlag("skip-registry-check", SetBool, nil, nil, true)
	PublicHostname    = createFlag("public-hostname", SetString, nil, nil, true)
	RoutingSuffix     = createFlag("routing-suffix", SetString, nil, nil, true)
	HostConfigDir     = createFlag("host-config-dir", SetString, []setFn{IsValidPath}, nil, true)
	HostVolumeDir     = createFlag("host-volumes-dir", SetString, []setFn{IsValidPath}, nil, true)
	HostDataDir       = createFlag("host-data-dir", SetString, []setFn{IsValidPath}, nil, true)
	HostPvDir         = createFlag("host-pv-dir", SetString, []setFn{IsValidPath}, nil, true)
	ServerLogLevel    = createFlag("server-loglevel", SetInt, []setFn{IsPositive}, nil, true)
	OpenshiftEnv      = createFlag("openshift-env", nil, nil, nil, false)
	Metrics           = createFlag("metrics", SetBool, nil, nil, true)
	Logging           = createFlag("logging", SetBool, nil, nil, true)
	ServiceCatalog    = createFlag("service-catalog", SetBool, nil, nil, true)
	// ability to specify additional cluster up flags
	OcClusterExtraFlags = createFlag("oc-cluster-extra-flags", SetString, nil, nil, true)

	// Setting proxy
	NoProxyList = createFlag("no-proxy", SetString, nil, nil, true)
	HttpProxy   = createFlag("http-proxy", SetString, []setFn{IsValidProxy}, nil, true)
	HttpsProxy  = createFlag("https-proxy", SetString, []setFn{IsValidProxy}, nil, true)

	// Subscription Manager
	Username         = createFlag("username", SetString, nil, nil, true)
	Password         = createFlag("password", SetString, nil, nil, true)
	SkipRegistration = createFlag("skip-registration", SetBool, nil, nil, true)

	// Global flags
	LogDir             = createFlag("log_dir", SetString, []setFn{IsValidPath}, nil, true)
	ShowLibmachineLogs = createFlag("show-libmachine-logs", SetBool, nil, nil, true)

	// Host Folders
	HostFoldersMountPath = createFlag("hostfolders-mountpath", SetString, nil, nil, true)
	HostFoldersAutoMount = createFlag("hostfolders-automount", SetBool, nil, nil, true)

	ImageCaching = createFlag("image-caching", SetBool, nil, nil, true)
)

func createFlag(name string, set func(MinishiftConfig, string, string) error, validations []setFn, callbacks []setFn, isApply bool) *Setting {
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
to the options that you set when you run the minishift start command.

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
		return nil, fmt.Errorf("Cannot open file %s: %s", constants.ConfigFile, err)
	}
	var m MinishiftConfig
	m, err = decode(f)
	if err != nil {
		return nil, fmt.Errorf("Cannot decode config %s: %s", constants.ConfigFile, err)
	}

	return m, nil
}

// Writes a config to the $MINISHIFT_HOME/config/config.json file
func WriteConfig(m MinishiftConfig) error {
	f, err := os.Create(constants.ConfigFile)
	if err != nil {
		return fmt.Errorf("Cannot create file %s: %s", constants.ConfigFile, err)
	}
	defer f.Close()
	err = encode(f, m)
	if err != nil {
		return fmt.Errorf("Cannot encode config %s: %s", constants.ConfigFile, err)
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
