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

	"bytes"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"sort"
	"strings"
)

const (
	DefaultConfigViewFormat = "- {{.ConfigKey | printf \"%-21s\"}}: {{.ConfigValue}}\n"
)

var configViewFormat string
var excludedConfigKeys = make(map[string]interface{})

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
			atexit.Exit(1)
		}
	},
}

func init() {
	excludedConfigKeys["addons"] = true
	configViewCmd.Flags().StringVar(&configViewFormat, "format", DefaultConfigViewFormat,
		`Go template format to apply to the configuration file. For more information about Go templates, see: https://golang.org/pkg/text/template/
		For the list of configurable variables for the template, see the struct values section of ConfigViewTemplate at: https://godoc.org/github.com/minishift/minishift/cmd/minishift/cmd/config#ConfigViewTemplate`)
	ConfigCmd.AddCommand(configViewCmd)
}

func configView() error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	for k, v := range cfg {
		_, excluded := excludedConfigKeys[k]
		if excluded {
			continue
		}

		tmpl, err := template.New("view").Parse(configViewFormat)
		if err != nil {
			glog.Errorln("Error creating view template:", err)
			atexit.Exit(1)
		}
		viewTmplt := ConfigViewTemplate{k, v}
		err = tmpl.Execute(&buffer, viewTmplt)
		if err != nil {
			glog.Errorln("Error executing view template:", err)
			atexit.Exit(1)
		}
	}

	lines := strings.Split(buffer.String(), "\n")
	// remove empy line at end
	lines = lines[:len(lines)-1]
	sort.Strings(lines)
	fmt.Print(strings.Join(lines, "\n"))
	return nil
}
