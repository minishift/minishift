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

package process

import (
	"fmt"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"os"
	"syscall"
)

func SysProcForBackgroundProcess() *syscall.SysProcAttr {
	sysProcAttr := new(syscall.SysProcAttr)
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms682425(v=vs.85).aspx
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms684863(v=vs.85).aspx
	// 0x00000010 = CREATE_NEW_CONSOLE
	sysProcAttr.CreationFlags = syscall.CREATE_NEW_PROCESS_GROUP | 0x00000010
	sysProcAttr.HideWindow = true

	return sysProcAttr
}

func EnvForBackgroundProcess() []string {
	return []string{
		fmt.Sprintf("MINISHIFT_HOME=%s", constants.Minipath),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("PATHEXT=%s", os.Getenv("PATHEXT")),
		fmt.Sprintf("SystemRoot=%s", os.Getenv("SystemRoot")),
		fmt.Sprintf("COMPUTERNAME=%s", os.Getenv("COMPUTERNAME")),
	}
}
