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
	"regexp"
	"strings"

	"path/filepath"

	"encoding/json"

	"github.com/docker/machine/libmachine/log"
	"github.com/golang/glog"
	"github.com/minishift/minishift/cmd/minishift/cmd/addon"
	configCmd "github.com/minishift/minishift/cmd/minishift/cmd/config"
	hostfolderCmd "github.com/minishift/minishift/cmd/minishift/cmd/hostfolder"
	"github.com/minishift/minishift/cmd/minishift/cmd/image"
	cmdOpenshift "github.com/minishift/minishift/cmd/minishift/cmd/openshift"
	cmdProfile "github.com/minishift/minishift/cmd/minishift/cmd/profile"
	"github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/minishift/minishift/pkg/version"
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
	constants.MakeMiniPath("cache", "images"),
	constants.MakeMiniPath("config"),
	constants.MakeMiniPath("addons"),
	constants.MakeMiniPath("logs"),
	constants.MakeMiniPath("tmp"),
}

const (
	showLibmachineLogs    = "show-libmachine-logs"
	profile               = "profile"
	enableExperimentalEnv = "MINISHIFT_ENABLE_EXPERIMENTAL"
)

var viperWhiteList = []string{
	"v",
	"alsologtostderr",
	"log_dir",
}

var hasEnabledExperimental bool = false

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "minishift",
	Short: "Minishift is a tool for application development in local OpenShift clusters.",
	Long:  `Minishift is a command-line tool that provisions and manages single-node OpenShift clusters optimized for development workflows.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var (
			err          error
			isFreshStart bool
		)

		//We need to initialize allinstance config as the profile information would be there
		//To do that we need to create the directories for it.
		for _, path := range dirs {
			if match, _ := regexp.MatchString("config", path); match {
				if err := os.MkdirAll(path, 0777); err != nil {
					atexit.ExitWithMessage(1, fmt.Sprintf("Error creating minishift directory: %s", err))
				}
			}
		}

		// Create all instances config
		minishiftConfig.AllInstancesConfig, err = minishiftConfig.NewAllInstancesConfig(constants.AllInstanceConfigPath)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error creating config for all instances: %s", err.Error()))
		}

		if cmdProfile.ProfileExists() {

			//Update MachineName
			constants.MachineName = cmdProfile.GetProfileName()
			constants.Minipath = constants.GetMinishiftHomeDir()

			//constants.Minipath := constants.Minipath
			fmt.Println("Minipath = ", constants.Minipath)

			constants.KubeConfigPath = filepath.Join(constants.Minipath, "machines", constants.MachineName+"_kubeconfig")

			//we have to recalculate the paths to Minishift directories except the config
			//As for config we are not doing anything at this point
			dirs[0] = constants.Minipath
			dirs[1] = constants.MakeMiniPath("certs")
			dirs[2] = constants.MakeMiniPath("machines")
			dirs[3] = constants.MakeMiniPath("cache")
			dirs[4] = constants.MakeMiniPath("cache", "iso")
			dirs[5] = constants.MakeMiniPath("cache", "oc")
			dirs[6] = constants.MakeMiniPath("cache", "images")
			dirs[8] = constants.MakeMiniPath("addons")
			dirs[9] = constants.MakeMiniPath("logs")
			dirs[10] = constants.MakeMiniPath("tmp")

		}

		if !filehelper.Exists(constants.Minipath) || filehelper.IsEmptyDir(constants.Minipath) {
			isFreshStart = true
		}

		//creating all directories for minishift run
		for _, path := range dirs {
			if err := os.MkdirAll(path, 0777); err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Error creating minishift directory: %s", err))
			}
		}

		ensureConfigFileExists(constants.ConfigFile)

		// Create MACHINE_NAME.json
		instanceConfigPath := filepath.Join(constants.Minipath, "machines", constants.MachineName+".json")
		minishiftConfig.InstanceConfig, err = minishiftConfig.NewInstanceConfig(instanceConfigPath)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error creating config for VM: %s", err.Error()))
		}

		// Run the default addons on fresh minishift start
		if isFreshStart {
			fmt.Print("-- Installing default add-ons ... ")
			if err := util.UnpackAddons(constants.MakeMiniPath("addons")); err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Error installing default add-ons : %s", err))
			}

			fmt.Println("OK")
		}

		// Check marker file created by update command and perform post update execu	tion steps
		if filehelper.Exists(filepath.Join(constants.Minipath, constants.UpdateMarkerFileName)) {
			if err := performPostUpdateExecution(filepath.Join(constants.Minipath, constants.UpdateMarkerFileName)); err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Error in performing post update exeuction: %s", err))
			}
		}

		if hasEnabledExperimental {
			glog.Info("Experimental features are enabled")
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
	enableExperimental, err := util.GetBoolEnv(enableExperimentalEnv)
	if err == util.BooleanFormatError {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error enabling experimental features: %s", err))
	}
	hasEnabledExperimental = enableExperimental

	RootCmd.PersistentFlags().Bool(showLibmachineLogs, false, "Show logs from libmachine.")
	RootCmd.PersistentFlags().String(profile, "", "Profile name")
	RootCmd.AddCommand(configCmd.ConfigCmd)
	RootCmd.AddCommand(cmdOpenshift.OpenShiftCmd)
	RootCmd.AddCommand(hostfolderCmd.HostfolderCmd)
	RootCmd.AddCommand(addon.AddonsCmd)
	RootCmd.AddCommand(image.ImageCmd)
	RootCmd.AddCommand(cmdProfile.ProfileCmd)
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

// performPostUpdateExecution executes the post update actions like unpacking the default addons
// if user chose to update addons during `minishift update` command.
// It also remove the marker file created by update command to avoid repeating the post update execution process
func performPostUpdateExecution(markerPath string) error {
	var markerData UpdateMarker

	file, err := ioutil.ReadFile(markerPath)
	if err != nil {
		return err
	}

	json.Unmarshal(file, &markerData)
	if markerData.InstallAddon {
		fmt.Println(fmt.Sprintf("Minishift was upgraded from v%s to v%s. Running post update actions.", markerData.PreviousVersion, version.GetMinishiftVersion()))
		fmt.Print("--- Updating default add-ons ... ")
		util.UnpackAddons(constants.MakeMiniPath("add-ons"))
		fmt.Println("OK")
		fmt.Println(fmt.Sprintf("Default add-ons %s installed", strings.Join(util.DefaultAssets, ", ")))
	}

	// Delete the marker file once post update execution is done
	if err := os.Remove(markerPath); err != nil {
		return err
	}

	return nil
}
