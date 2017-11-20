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
	"reflect"
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
	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	profileActions "github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/minishift/minishift/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	showLibmachineLogs    = "show-libmachine-logs"
	profileFlag           = "profile"
	enableExperimentalEnv = "MINISHIFT_ENABLE_EXPERIMENTAL"
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
		var (
			err                    error
			isAddonInstallRequired bool
		)

		constants.MachineName = constants.ProfileName

		// Initialize the instance directory structure
		minishiftConfig.InstanceDirs = minishiftConfig.NewMinishiftDirs()

		constants.KubeConfigPath = filepath.Join(constants.Minipath, "machines", constants.MachineName+"_kubeconfig")

		if !filehelper.Exists(minishiftConfig.InstanceDirs.Addons) {
			isAddonInstallRequired = true
		}

		// creating all directories for minishift run
		createMinishiftDirs(minishiftConfig.InstanceDirs)

		// If AllInstanceConfig is not defined we should define it now.
		if minishiftConfig.AllInstancesConfig == nil {
			ensureAllInstanceConfigPath(constants.AllInstanceConfigPath)
			minishiftConfig.AllInstancesConfig, err = minishiftConfig.NewAllInstancesConfig(constants.AllInstanceConfigPath)
			if err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Error creating all instance config: %s", err.Error()))
			}
		}

		ensureConfigFileExists(constants.ConfigFile)

		// Create MACHINE_NAME.json
		instanceConfigPath := filepath.Join(constants.Minipath, "machines", constants.MachineName+".json")
		minishiftConfig.InstanceConfig, err = minishiftConfig.NewInstanceConfig(instanceConfigPath)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error creating config for VM: %s", err.Error()))
		}

		if isAddonInstallRequired {
			if err := cmdUtil.UnpackAddons(minishiftConfig.InstanceDirs.Addons); err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Error installing default add-ons : %s", err))
			}
		}

		// Check marker file created by update command and perform post update execution steps
		if filehelper.Exists(filepath.Join(constants.Minipath, constants.UpdateMarkerFileName)) {
			if err := performPostUpdateExecution(filepath.Join(constants.Minipath, constants.UpdateMarkerFileName)); err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Error in performing post update exeuction: %s", err))
			}
		}

		if minishiftConfig.EnableExperimental {
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

		setDefaultActiveProfile()
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

func processEnvVariables() {
	enableExperimental, err := cmdUtil.GetBoolEnv(enableExperimentalEnv)
	if err == cmdUtil.BooleanFormatError {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error enabling experimental features: %s", err))
	}

	minishiftConfig.EnableExperimental = enableExperimental
}

func init() {
	processEnvVariables()
	RootCmd.PersistentFlags().Bool(showLibmachineLogs, false, "Show logs from libmachine.")
	RootCmd.PersistentFlags().String(profileFlag, constants.DefaultProfileName, "Profile name")
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
	profile := initializeProfile()
	if (profile != "") && (profile != constants.DefaultProfileName) {
		constants.ProfileName = profile
		constants.ConfigFile = constants.MakeMiniPath("profiles", profile, "config", "config.json")
	}
	configPath := constants.ConfigFile
	viper.SetConfigFile(configPath)
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil {
		glog.Warningf("Error reading config file at %s: %s", configPath, err)
	}
	setupViper()
}

func initializeProfile() string {
	var (
		profileName   string
		err           error
		activeProfile string
	)

	for i, arg := range os.Args {
		if arg == "--"+profileFlag {
			profileName = os.Args[i+1]
			if !cmdUtil.IsValidProfileName(profileName) {
				atexit.ExitWithMessage(1, "Profile names must consist of alphanumeric characters only.")
			}
			break
		}
	}

	// Check if the allinstance config is present. If present we need to check active profile information.
	_, err = os.Stat(constants.AllInstanceConfigPath)
	if !os.IsNotExist(err) {
		minishiftConfig.AllInstancesConfig, err = minishiftConfig.NewAllInstancesConfig(constants.AllInstanceConfigPath)
		if err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error initializing all instance config: %s", err.Error()))
		}
		activeProfile = profileActions.GetActiveProfile()
	}

	if profileName != "" {
		return profileName
	}
	if activeProfile != "" {
		return activeProfile
	}
	return ""
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
		cmdUtil.UnpackAddons(minishiftConfig.InstanceDirs.Addons)
		fmt.Println("OK")
		fmt.Println(fmt.Sprintf("Default add-ons %s installed", strings.Join(cmdUtil.DefaultAssets, ", ")))
	}

	// Delete the marker file once post update execution is done
	if err := os.Remove(markerPath); err != nil {
		return err
	}

	return nil
}

func ensureAllInstanceConfigPath(configPath string) {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0777); err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating directory: %s", configDir))
	}
}

func createMinishiftDirs(dirs *minishiftConfig.MinishiftDirs) {
	dirPaths := reflect.ValueOf(*dirs)

	for i := 0; i < dirPaths.NumField(); i++ {
		path := dirPaths.Field(i).Interface().(string)
		if err := os.MkdirAll(path, 0777); err != nil {
			atexit.ExitWithMessage(1, fmt.Sprintf("Error creating directory: %s", path))
		}
	}
}

// If there is no active profile we need to set minishift as the default profile.
// Because this will make the profile behaviour backward compatible and consistent with user expectation.
func setDefaultActiveProfile() {
	if minishiftConfig.AllInstancesConfig == nil {
		atexit.ExitWithMessage(1, "Error: All instance config is not initialized")
	}
	activeProfile := profileActions.GetActiveProfile()
	if activeProfile == "" {
		err := profileActions.SetDefaultProfileActive()
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}

		// Only set oc context to default profile when user is looking for default profile
		// i.e. "minishift start" with minishift as active profile or "minishift start --profile minishift"
		// Otherwise minishift will be the active profile irrespective of what user chooses
		if constants.ProfileName == constants.DefaultProfileName {
			cmdUtil.SetOcContext(constants.DefaultProfileName)
		}
	}
}
