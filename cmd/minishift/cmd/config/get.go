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
	"fmt"
	"os"

	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var configGetCmd = &cobra.Command{
	Use:   "get PROPERTY_NAME",
	Short: "Gets the value of a configuration property from the Minishift configuration file.",
	Long:  "Gets the value of a configuration property from the minishift configuration file. This value can be overwritten at runtime by flags or environmental variables.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			atexit.ExitWithMessage(1, "usage: minishift config get PROPERTY_NAME")
		}

		val, err := get(args[0])
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
		}
		if val != "" {
			fmt.Fprintln(os.Stdout, val)
		}
	},
}

func init() {
	ConfigCmd.AddCommand(configGetCmd)
}

func get(name string) (string, error) {
	m, err := ReadConfig()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", m[name]), nil
}
