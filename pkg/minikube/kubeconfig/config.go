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

package kubeconfig

import (
	"fmt"
	"github.com/docker/machine/libmachine/provision"
	"io/ioutil"
)

// Cache system admin entries to be used to run oc commands
func CacheSystemAdminEntries(systemEntriesConfigPath string, ocPath string, sshCommander provision.SSHCommander) error {
	// There is another easy way to get config for current context
	// oc login -u system:admin
	// oc config view --minify --raw=true
	// We need to login as system:admin because then only config view will have client-certificate-data
	// and client-key-data which is associated with admin and all the operation we do as oc runner.
	cmd := fmt.Sprintf("sudo %s login -u system:admin", ocPath)
	_, err := sshCommander.SSHCommand(cmd)
	if err != nil {
		return err
	}

	cmd = fmt.Sprintf("sudo %s config view --minify --raw=true", ocPath)
	out, err := sshCommander.SSHCommand(cmd)
	if err != nil {
		return err
	}

	// Write to machines/<MACHINE_NAME>_kubeconfig
	if err := ioutil.WriteFile(systemEntriesConfigPath, []byte(out), 0644); err != nil {
		return err
	}

	return nil
}
