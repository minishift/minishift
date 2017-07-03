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

package image

import (
	"github.com/spf13/cobra"
)

const (
	noImageSpecified = "You need to specify one or more images."
)

var ImageCacheConfigCmd = &cobra.Command{
	Use:   "cache-config SUBCOMMAND [flags]",
	Short: "Controls the list of cached images which are implicitly imported and exported.",
	Long:  "Controls the list of cached images which are implicitly imported and exported.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	ImageCmd.AddCommand(ImageCacheConfigCmd)
}
