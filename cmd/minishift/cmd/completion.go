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
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const longDescription = `
	Outputs minishift shell completion for the given shell (bash or zsh)

	This depends on the bash-completion binary.  Example installation instructions:

	macOS:

		$ brew install bash-completion
		$ source $(brew --prefix)/etc/bash_completion
		$ minishift completion bash > ~/.minishift-completion	# for bash users
		$ minishift completion zsh > ~/.minishift-completion	# for zsh users
		$ source ~/.minishift-completion

	RHEL/Fedora:

		$ yum install bash-completion				# for RHEL
		$ dnf install bash-completion   			# for Fedora
		$ minishift completion bash > ~/.minishift-completion 	# for bash users
		$ minishift completion zsh > ~/.minishift-completion	# for zsh users
		$ source ~/.minishift-completion

	Additionally, you may want to output the completion to a file and source in your .bashrc

	Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2
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
	Short: "Outputs minishift shell completion for the given shell (bash or zsh)",
	Long:  longDescription,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// NOOP
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Usage: minishift completion SHELL")
			atexit.Exit(1)
		}
		if args[0] != "bash" && args[0] != "zsh" {
			fmt.Println("Only bash and zsh are supported for minishift completion")
			atexit.Exit(1)
		} else if args[0] == "bash" {
			err := GenerateBashCompletion(os.Stdout, cmd.Parent())
			if err != nil {
				atexit.ExitWithMessage(1, err.Error())
			}
		} else {
			err := GenerateZshCompletion(os.Stdout, cmd.Parent())
			if err != nil {
				atexit.ExitWithMessage(1, err.Error())
			}
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

func GenerateZshCompletion(out io.Writer, cmd *cobra.Command) error {
	zsh_initialization := `
__minishift_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__minishift_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__minishift_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__minishift_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__minishift_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}
__minishift_declare() {
	if [ "$1" == "-F" ]; then
		whence -w "$@"
	else
		builtin declare "$@"
	fi
}
__minishift_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__minishift_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__minishift_filedir() {
	local RET OLD_IFS w qw
	__debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS="," __debug "RET=${RET[@]} len=${#RET[@]}"
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__minishift_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}
__minishift_quote() {
	if [[ $1 == \'* || $1 == \"* ]]; then
		# Leave out first character
		printf %q "${1:1}"
	else
		printf %q "$1"
	fi
}
autoload -U +X bashcompinit && bashcompinit
# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi
__minishift_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__minishift_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__minishift_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__minishift_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__minishift_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__minishift_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/__minishift_declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__minishift_type/g" \
	<<'BASH_COMPLETION_EOF'
`

	_, err := out.Write([]byte(boilerPlate))
	if err != nil {
		return err
	}

	_, err = out.Write([]byte(zsh_initialization))
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = cmd.GenBashCompletion(buf)
	if err != nil {
		return errors.Wrap(err, "Error generating zsh completion")
	}
	_, err = out.Write(buf.Bytes())
	if err != nil {
		return err
	}

	zsh_tail := `
BASH_COMPLETION_EOF
}
__minishift_bash_source <(__minishift_convert_bash_to_zsh)
`
	_, err = out.Write([]byte(zsh_tail))
	if err != nil {
		return err
	}

	return nil
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
