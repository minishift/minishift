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
	"io"
	"os"

	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const longDescription = `
	Outputs minishift shell completion for the given shell (bash)

	This depends on the bash-completion binary.  Example installation instructions:
	OS X:
		$ brew install bash-completion
		$ source $(brew --prefix)/etc/bash_completion
		$ minishift completion bash > ~/.minishift-completion
		$ source ~/.minishift-completion

	Fedora:
		$ sudo dnf install bash-completion
		$ minishift completion bash > ~/.minishift-completion
		$ source ~/.minishift-completion

	Ubuntu:
		$ apt-get install bash-completion
		$ source /etc/bash-completion
		$ source <(minishift completion bash)

	Additionally, you may want to output completion to a file and source in your .bashrc
`

const boilerPlate = `
# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
`

var completionCmd = &cobra.Command{
	Use:   "completion SHELL",
	Short: "Outputs minishift shell completion for the given shell (bash)",
	Long:  longDescription,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Usage: minishift completion SHELL")
			atexit.Exit(1)
		}
		if args[0] != "bash" {
			fmt.Println("Only bash is supported for minishift completion")
			atexit.Exit(1)
		}
		err := GenerateBashCompletion(os.Stdout, cmd.Parent())
		if err != nil {
			fmt.Println(fmt.Sprintf("%s", err.Error()))
			atexit.Exit(1)
		}
	},
}

func GenerateBashCompletion(w io.Writer, cmd *cobra.Command) error {
	_, err := w.Write([]byte(boilerPlate))
	if err != nil {
		return err
	}

	err = cmd.GenBashCompletion(w)
	if err != nil {
		return errors.Wrap(err, "Error generating bash completion")
	}

	return nil
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
