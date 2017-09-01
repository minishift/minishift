/*
Copyright (C) 2017 Red Hat, Inc.

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

package profile

import (
	"github.com/spf13/cobra"
)

const (
	emptyProfileError  = "You must provide the profile name. Run `minishift profile list` to view profiles"
	extraArgumentError = "You have provided more arguments than required. You must provide a single profile name"
)

var ProfileCmd = &cobra.Command{
	Use:     "profile SUBCOMMAND [flags]",
	Aliases: []string{"instance"},
	Short:   "Manages Minishift profiles.",
	Long:    "Allows you to create and manage multiple Minishift instances (profiles). Use the sub-commands to set active and list existing profiles.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
