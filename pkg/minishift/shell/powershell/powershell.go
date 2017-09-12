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

package powershell

import (
	ps "github.com/gorillalabs/go-powershell"
	"github.com/gorillalabs/go-powershell/backend"
)

type PowerShell struct {
	powerShell ps.Shell
}

func New() *PowerShell {
	return &PowerShell{
		powerShell: createPowerShell(),
	}
}
func (p *PowerShell) Execute(command string) (stdOut string, stdErr string) {
	stdOut, stdErr, _ = p.powerShell.Execute(command)
	return
}

func createPowerShell() ps.Shell {
	back := &backend.Local{}
	shell, _ := ps.New(back)
	return shell
}
