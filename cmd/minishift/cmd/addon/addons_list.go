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
	"text/tabwriter"
	"text/template"

	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/minishift/minishift/pkg/minishift/addon"
	"github.com/minishift/minishift/pkg/minishift/addon/manager"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/cobra"
)

var verbose bool

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
	Short: "Lists all installed Minishift add-ons.",
	Long:  "Lists all installed Minishift add-ons and their current status, such as enabled/disabled.",
	Run:   runListCommand,
}

func init() {
	addonsListCmd.Flags().BoolVar(&verbose, "verbose", false, "Prints the add-on list with a more verbose format of the output that includes the add-on description.")
	AddonsCmd.AddCommand(addonsListCmd)
}

func runListCommand(cmd *cobra.Command, args []string) {
	addOnManager := GetAddOnManager()
	verboseListTemplate, err := template.New("list").Parse(verboseAddonListFormat)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error creating the list template: %s", err.Error()))
	}
	printAddOnList(addOnManager, os.Stdout, verboseListTemplate)
}

func printAddOnList(manager *manager.AddOnManager, writer io.Writer, template *template.Template) {
	addOns := manager.List()
	sort.Sort(addon.ByStatusThenPriorityThenName(addOns))
	display := new(tabwriter.Writer)
	display.Init(os.Stdout, 0, 8, 0, '\t', 0)

	for _, addon := range addOns {
		description := strings.Join(addon.MetaData().Description(), fmt.Sprintf("\n%13s", " "))
		addonInfo := DisplayAddOn{addon.MetaData().Name(), description, stringFromStatus(addon.IsEnabled()), addon.GetPriority()}
		if verbose {
			err := template.Execute(writer, addonInfo)
			if err != nil {
				atexit.ExitWithMessage(1, fmt.Sprintf("Error executing the template: %s", err.Error()))
			}
		} else {
			fmt.Fprintln(display, fmt.Sprintf("- %s\t : %s\tP(%v)", addonInfo.Name, addonInfo.Status, addonInfo.Priority))
		}
	}
	display.Flush()
}

func stringFromStatus(addonStatus bool) string {
	if addonStatus {
		if verbose {
			return "true"
		}
		return "enabled"
	}
	if verbose {
		return "false"
	}
	return "disabled"
}
