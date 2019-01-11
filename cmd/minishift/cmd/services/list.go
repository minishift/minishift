/*
Copyright (C) 2018 Red Hat, Inc.

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

package services

import (
	"fmt"

	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/spf13/cobra"
)

// serviceListCmd represents the list command
var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the available Minishift services.",
	Long:  "List the available Minishift services.",
	Run:   runServiceList,
}

func init() {
	ServicesCmd.AddCommand(serviceListCmd)
}

func runServiceList(cmd *cobra.Command, args []string) {
	fmt.Printf("The following Minishift services are available: \n")
	for _, component := range minishiftConstants.ValidServices {
		fmt.Printf("\t- %s\n", component)
	}
}
