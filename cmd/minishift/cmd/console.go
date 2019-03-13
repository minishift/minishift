/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"os"

	"github.com/docker/machine/libmachine"
	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var (
	consoleURLMode    bool
	machineReadAble   bool
	requestOauthToken bool
	machineDetails    = `HOST=%s
PORT=%d
CONSOLE_URL=%s`
)

// consoleCmd represents the console command
var consoleCmd = &cobra.Command{
	Use:     "console",
	Aliases: []string{"dashboard"},
	Short:   "Opens or displays the OpenShift Web Console URL.",
	Long:    `Opens the OpenShift Web Console URL in the default browser or displays it to the console.`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
		defer api.Close()

		if consoleURLMode {
			fmt.Fprintln(os.Stdout, getHostUrl(api))
		} else if machineReadAble {
			displayConsoleInMachineReadable(getHostIp(api), getHostUrl(api))
		} else if requestOauthToken {
			fmt.Fprintln(os.Stdout, "Opening requested token URI in the default browser...")
			browser.OpenURL(getTokenRequestUrl(api))
		} else {
			fmt.Fprintln(os.Stdout, "Opening the OpenShift Web console in the default browser...")
			browser.OpenURL(getHostUrl(api))
		}
	},
}

func displayConsoleInMachineReadable(hostIP string, url string) {
	machineDetails = fmt.Sprintf(machineDetails, hostIP, constants.APIServerPort, url)
	fmt.Fprintln(os.Stdout, machineDetails)
}

func getHostUrl(api *libmachine.Client) string {
	host, err := api.Load(constants.MachineName)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Error getting Minishift status: %s", err.Error()))
	}
	cmdUtil.ExitIfNotRunning(host.Driver, constants.MachineName)

	url, err := cluster.GetConsoleURL(api)
	if err != nil {
		atexit.ExitWithMessage(1, fmt.Sprintf("Cannot access the OpenShift console. Verify that Minishift is running. Error: %s", err.Error()))
	}
	return url
}

func getTokenRequestUrl(api *libmachine.Client) string {
	hostIP := getHostIp(api)
	tokenUrl := fmt.Sprintf("https://%s:%d/%s", hostIP, constants.APIServerPort, "oauth/token/request")
	return tokenUrl
}

func getHostIp(api *libmachine.Client) string {
	hostIP, err := cluster.GetHostIP(api)
	if err != nil {
		fmt.Println("Cannot get Host IP. Verify that Minishift is running. Error: ", err)
	}
	return hostIP
}

func init() {
	consoleCmd.Flags().BoolVar(&consoleURLMode, "url", false, "Prints the OpenShift Web Console URL to the console.")
	consoleCmd.Flags().BoolVar(&machineReadAble, "machine-readable", false, "Prints OpenShift's IP, port and Web Console URL in Machine readable format")
	consoleCmd.Flags().BoolVar(&requestOauthToken, "request-oauth-token", false, "Open token request to default web browser")
	RootCmd.AddCommand(consoleCmd)
}
