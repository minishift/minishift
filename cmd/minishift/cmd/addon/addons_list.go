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
	"os"
	"text/template"

	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/minishift/addon"
	"github.com/minishift/minishift/pkg/minishift/addon/manager"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
	"io"
	"sort"
)

var verbose bool

var defaultAddonListFormat = "- {{.Name | printf \"%-15s\"}}: {{.Status | printf \"%-10s\"}} P({{.Priority | printf \"%d\"}})\n"
var defaultListTemplate *template.Template

var verboseAddonListFormat = `Name       : {{.Name}}
Description: {{.Description}}
Enabled    : {{.Status}}
Priority   : {{.Priority}}

`
var verboseListTemplate *template.Template

type DisplayAddOn struct {
	Name        string
	Description string
	Status      string
	Priority    int
}

var addonsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all installed Minishift addons",
	Long:  "Lists all installed Minishift addons as well as there current status (enabled/disabled)",
	Run:   runListCommand,
}

func init() {
	addonsListCmd.Flags().BoolVar(&verbose, "verbose", false, "A more verbose format of the output including the addon description.")
	AddonsCmd.AddCommand(addonsListCmd)

	var err error
	defaultListTemplate, err = template.New("list").Parse(defaultAddonListFormat)
	if err != nil {
		glog.Errorln("Error creating list template:", err)
		atexit.Exit(1)
	}

	verboseListTemplate, err = template.New("list").Parse(verboseAddonListFormat)
	if err != nil {
		glog.Errorln("Error creating list template:", err)
		atexit.Exit(1)
	}
}

func runListCommand(cmd *cobra.Command, args []string) {
	addOnManager := GetAddOnManager()

	template := defaultListTemplate
	if verbose {
		template = verboseListTemplate
	}

	printAddOnList(addOnManager, os.Stdout, template)
}

func printAddOnList(manager *manager.AddOnManager, writer io.Writer, template *template.Template) {
	addOns := manager.List()
	sort.Sort(addon.ByStatusThenPriorityThenName(addOns))
	for _, addon := range addOns {
		addonTemplate := DisplayAddOn{addon.MetaData().Name(), addon.MetaData().Description(), stringFromStatus(addon.IsEnabled()), addon.GetPriority()}
		err := template.Execute(writer, addonTemplate)
		if err != nil {
			glog.Errorln("Error executing template: ", err)
			atexit.Exit(1)
		}
	}
}

func stringFromStatus(addonStatus bool) string {
	if addonStatus {
		return "enabled"
	}
	return "disabled"
}
