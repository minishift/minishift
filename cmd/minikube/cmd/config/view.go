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
	"text/template"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"github.com/minishift/minishift/pkg/minikube/constants"
)

var configViewFormat string

type ConfigViewTemplate struct {
	ConfigKey   string
	ConfigValue interface{}
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Display the properties and values of the Minishift configuration file.",
	Long:  "Display the properties and values of the Minishift configuration file. You can set the output format from one of the available Go templates.",
	Run: func(cmd *cobra.Command, args []string) {
		err := configView()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	configViewCmd.Flags().StringVar(&configViewFormat, "format", constants.DefaultConfigViewFormat,
		`Go template format to apply to the configuration file. For more information about Go templates, see: https://golang.org/pkg/text/template/
		For the list of configurable variables for the template, see the struct values section of ConfigViewTemplate at: https://godoc.org/github.com/minishift/minishift/cmd/minikube/cmd/config#ConfigViewTemplate`)
}

func init() {
	ConfigCmd.AddCommand(configViewCmd)
}

func configView() error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	for k, v := range cfg {
		tmpl, err := template.New("view").Parse(configViewFormat)
		if err != nil {
			glog.Errorln("Error creating view template:", err)
			os.Exit(1)
		}
		viewTmplt := ConfigViewTemplate{k, v}
		err = tmpl.Execute(os.Stdout, viewTmplt)
		if err != nil {
			glog.Errorln("Error executing view template:", err)
			os.Exit(1)
		}
	}
	return nil
}
