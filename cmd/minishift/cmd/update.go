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

package cmd

import (
	"os"

	"github.com/minishift/minishift/pkg/minikube/update"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Checks for updates in Minishift",
	Long:  `Checks for updates in Minishift and prompts the user to download newer version`,
	Run:   runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) {
	update.MaybeUpdateFromGithub(os.Stdout)
}

func init() {
	RootCmd.AddCommand(updateCmd)
}
