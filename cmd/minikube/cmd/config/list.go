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

	"github.com/spf13/cobra"
)

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists the contents of the current minishift config file",
	Long:  "Lists the contents of the current minishift config file.",
	Run: func(cmd *cobra.Command, args []string) {
		err := printList()
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
		}
	},
}

func init() {
	ConfigCmd.AddCommand(configListCmd)
}

func printList() error {
	m, err := ReadConfig()
	if err != nil {
		return err
	}
	for name, value := range m {
		fmt.Printf("- %s: %v\n", name, value)
	}
	return nil
}
