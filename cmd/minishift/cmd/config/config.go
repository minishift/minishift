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

	"github.com/minishift/minishift/pkg/minikube/config"
	"github.com/minishift/minishift/pkg/minikube/constants"
)

type configFile interface {
	io.ReadWriter
}

type setFn func(string, string) error

type MinishiftConfig map[string]interface{}

type Setting struct {
	name        string
	set         func(MinishiftConfig, string, string) error
	validations []setFn
	callbacks   []setFn
}

// These are all the settings that are configurable
// and their validation and callback fn run on Set
var settings []Setting = []Setting{
	{
		name:        "vm-driver",
		set:         SetString,
		validations: []setFn{IsValidDriver},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "v",
		set:         SetInt,
		validations: []setFn{IsPositive},
	},
	{
		name:        "cpus",
		set:         SetInt,
		validations: []setFn{IsPositive},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "disk-size",
		set:         SetString,
		validations: []setFn{IsValidDiskSize},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name:        "host-only-cidr",
		set:         SetString,
		validations: []setFn{IsValidCIDR},
	},
	{
		name:        "memory",
		set:         SetInt,
		validations: []setFn{IsPositive},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name: "show-libmachine-logs",
		set:  SetBool,
	},
	{
		name:        "log_dir",
		set:         SetString,
		validations: []setFn{IsValidPath},
	},
	{
		name: "openshift-version",
		set:  SetString,
	},
	{
		name:        "iso-url",
		set:         SetString,
		validations: []setFn{IsValidUrl},
		callbacks:   []setFn{RequiresRestartMsg},
	},
	{
		name: "hostfolders-automount",
		set:  SetBool,
	},
	{
		name: "hostfolders-mountpath",
		set:  SetString,
	},
	{
		name: config.WantUpdateNotification,
		set:  SetBool,
	},
	{
		name: config.ReminderWaitPeriodInHours,
		set:  SetInt,
	},
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
	for _, s := range settings {
		fields = append(fields, " * "+s.name)
	}
	return strings.Join(fields, "\n")
}

// ReadConfig reads in the JSON minishift config
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

// Writes a minikube config to the JSON file
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
