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

package addon

import (
	"github.com/spf13/cobra"
)

var AddonsCmd = &cobra.Command{
	Use:   "addons SUBCOMMAND [flags]",
	Short: "Manages Minishift add-ons",
	Long:  "Manages Minishift add-ons. You can install, list, enable or disable Minishift add-ons.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
