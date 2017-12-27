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
	"strings"

	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

const (
	emptyProfileMessage  = "A profile name must be provided. Run 'minishift profile list' for a list of existing profiles."
	extraArgumentMessage = "You have provided more arguments than required. You must provide a single profile name."
	invalidNameMessage   = "Profile names must consist of alphanumeric characters only."
)

var ProfileCmd = &cobra.Command{
	Use:     "profile SUBCOMMAND [flags]",
	Aliases: []string{"instance", "profiles"},
	Short:   "Manages Minishift profiles.",
	Long:    "Allows you to create and manage multiple Minishift instances. Use the sub-commands to list, activate or delete existing profiles.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func validateArgs(args []string) {
	if len(args) == 0 {
		atexit.ExitWithMessage(1, emptyProfileMessage)
	} else if len(args) > 1 {
		atexit.ExitWithMessage(1, extraArgumentMessage)
	} else if strings.TrimSpace(args[0]) == "" {
		atexit.ExitWithMessage(1, emptyProfileMessage)
	}
	if !cmdUtil.IsValidProfileName(args[0]) {
		atexit.ExitWithMessage(1, invalidNameMessage)
	}
}
