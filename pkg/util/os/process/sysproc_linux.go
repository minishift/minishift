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
	sysProcAttr.Setpgid = true
	sysProcAttr.Pgid = 0

	return sysProcAttr
}

func EnvForBackgroundProcess() []string {
	return []string{
		fmt.Sprintf("MINISHIFT_HOME=%s", constants.Minipath),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}
}
