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
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"text/template"

	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

const (
	DefaultConfigViewFormat = "- {{.ConfigKey | printf \"%-21s\"}}: {{.ConfigValue}}"
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
		cfg, err := ReadConfig()
		if err != nil {
			atexit.ExitWithMessage(1, err.Error())
		}
		template := determineTemplate(configViewFormat)
		if err = configView(cfg, template, os.Stdout); err != nil {
			atexit.ExitWithMessage(1, err.Error())
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

func determineTemplate(tempFormat string) (tmpl *template.Template) {
	tmpl, err := template.New("view").Parse(tempFormat)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating view template: %s", err.Error()))
	}
	return tmpl
}

func configView(cfg MinishiftConfig, tmpl *template.Template, writer io.Writer) error {
	var lines []string
	for k, v := range cfg {
		_, excluded := excludedConfigKeys[k]
		if excluded {
			continue
		}
		viewTmplt := ConfigViewTemplate{k, v}
		var buffer bytes.Buffer
		if err := tmpl.Execute(&buffer, viewTmplt); err != nil {
			return err
		}
		lines = append(lines, buffer.String())
	}
	sort.Strings(lines)

	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}

	return nil
}
