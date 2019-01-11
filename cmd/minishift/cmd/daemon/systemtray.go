// +build !systemtray

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

package daemon

import (
	"runtime"

	"github.com/anjannath/systray"
	"github.com/minishift/minishift/pkg/minishift/systemtray"
	"github.com/spf13/cobra"
)

// systemtrayServiceCmd represents the systemtray command
var systemtrayDaemonCmd = &cobra.Command{
	Use:   "systemtray",
	Short: "Run minishift systemtray.",
	Long:  "Run a systemtray in the notification area of top bar/start menu.",
	Run: func(cmd *cobra.Command, args []string) {
		systray.Run(systemtray.OnReady, systemtray.OnExit)
	},
	Hidden: true,
}

func init() {
	if runtime.GOOS != "linux" {
		DaemonCmd.AddCommand(systemtrayDaemonCmd)
	}
}
