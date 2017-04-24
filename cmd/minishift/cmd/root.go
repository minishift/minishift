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
	goflag "flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"path/filepath"

	"github.com/docker/machine/libmachine/log"
	"github.com/golang/glog"
	"github.com/minishift/minishift/cmd/minishift/cmd/addon"
	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	hostfolderCmd "github.com/minishift/minishift/cmd/minishift/cmd/hostfolder"
	cmdOpenshift "github.com/minishift/minishift/cmd/minishift/cmd/openshift"
	"github.com/minishift/minishift/pkg/minikube/config"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var dirs = [...]string{
	constants.Minipath,
	constants.MakeMiniPath("certs"),
	constants.MakeMiniPath("machines"),
	constants.MakeMiniPath("cache"),
	constants.MakeMiniPath("cache", "iso"),
	constants.MakeMiniPath("cache", "oc"),
	constants.MakeMiniPath("config"),
	constants.MakeMiniPath("addons"),
	constants.MakeMiniPath("logs"),
	constants.MakeMiniPath("tmp"),
}

const (
	showLibmachineLogs = "show-libmachine-logs"
)

var viperWhiteList = []string{
	"v",
	"alsologtostderr",
	"log_dir",
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "minishift",
	Short: "Minishift is a tool for application development in local OpenShift clusters.",
	Long:  `Minishift is a command-line tool that provisions and manages single-node OpenShift clusters optimized for development workflows.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		for _, path := range dirs {
			if err := os.MkdirAll(path, 0777); err != nil {
				glog.Exitf("Error creating minishift directory: %s", err)
			}
		}

		ensureConfigFileExists(constants.ConfigFile)

		// Create all instances config
		allInstanceConfigPath := filepath.Join(constants.Minipath, "config", "allinstances.json")
		minishiftConfig.AllInstancesConfig, err = minishiftConfig.NewAllInstancesConfig(allInstanceConfigPath)

		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error creating config for all instances: %s", err.Error()))
		}

		// Create MACHINE_NAME.json
		instanceConfigPath := filepath.Join(constants.Minipath, "machines", constants.MachineName+".json")
		minishiftConfig.InstanceConfig, err = minishiftConfig.NewInstanceConfig(instanceConfigPath)

		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error creating config for VM: %s", err.Error()))
		}

		shouldShowLibmachineLogs := viper.GetBool(showLibmachineLogs)
		if glog.V(3) {
			log.SetDebug(true)
		}
		if !shouldShowLibmachineLogs {
			log.SetOutWriter(ioutil.Discard)
			log.SetErrWriter(ioutil.Discard)
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}
}

// Handle config values for flags used in external packages (e.g. glog)
// by setting them directly, using values from viper when not passed in as args
func setFlagsUsingViper() {
	for _, config := range viperWhiteList {
		var a = pflag.Lookup(config)
		viper.SetDefault(a.Name, a.DefValue)
		// If the flag is set, override viper value
		if a.Changed {
			viper.Set(a.Name, a.Value.String())
		}
		// Viper will give precedence first to calls to the Set command,
		// then to values from the config.yml
		a.Value.Set(viper.GetString(a.Name))
		a.Changed = true
	}
}

func init() {
	RootCmd.PersistentFlags().Bool(showLibmachineLogs, false, "Show logs from libmachine.")
	RootCmd.AddCommand(configCmd.ConfigCmd)
	RootCmd.AddCommand(cmdOpenshift.OpenShiftCmd)
	RootCmd.AddCommand(hostfolderCmd.HostfolderCmd)
	RootCmd.AddCommand(addon.AddonsCmd)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	logDir := pflag.Lookup("log_dir")
	if !logDir.Changed {
		logDir.Value.Set(constants.MakeMiniPath("logs"))
	}
	viper.BindPFlags(RootCmd.PersistentFlags())
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configPath := constants.ConfigFile
	viper.SetConfigFile(configPath)
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil {
		glog.Warningf("Error reading config file at %s: %s", configPath, err)
	}
	setupViper()
}

func setupViper() {
	viper.SetEnvPrefix(constants.MiniShiftEnvPrefix)
	// Replaces '-' in flags with '_' in env variables
	// e.g. show-libmachine-logs => $ENVPREFIX_SHOW_LIBMACHINE_LOGS
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetDefault(config.WantUpdateNotification, true)
	viper.SetDefault(config.ReminderWaitPeriodInHours, 24)
	setFlagsUsingViper()
}

func ensureConfigFileExists(configPath string) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		jsonRoot := []byte("{}")
		f, err := os.Create(configPath)
		if err != nil {
			glog.Exitf("Cannot create file %s: %s", configPath, err)
		}
		defer f.Close()
		_, err = f.Write(jsonRoot)
		if err != nil {
			glog.Exitf("Cannot encode config %s: %s", configPath, err)
		}
	}
}
