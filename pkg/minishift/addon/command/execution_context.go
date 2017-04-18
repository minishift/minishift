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

package command

import (
	"github.com/docker/machine/libmachine/provision"
	"github.com/minishift/minishift/pkg/minishift/docker"
	"github.com/minishift/minishift/pkg/minishift/oc"
)

// ExecutionContext contains the mapping of supported addon variables to their values
// as well as access to ssh and other needed resources to execute commands.
type ExecutionContext struct {
	InterpolationContext

	ocRunner             *oc.OcRunner
	dockerCommander      docker.DockerCommander
	sshCommander         provision.SSHCommander
	interpolationContext InterpolationContext
}

// NewExecutionContext creates a new execution context to be used with addon execution
func NewExecutionContext(ocPath string, kubeConfigPath string, sshCommander provision.SSHCommander) (*ExecutionContext, error) {
	ocRunner, err := oc.NewOcRunner(ocPath, kubeConfigPath)
	if err != nil {
		return nil, err
	}

	dockerCommander := docker.NewVmDockerCommander(sshCommander)
	context := NewInterpolationContext()
	return &ExecutionContext{ocRunner: ocRunner, dockerCommander: dockerCommander, sshCommander: sshCommander, interpolationContext: context}, nil
}

// GetSSHCommander returns a ssh commander to execute ssh commands against the Minishift VM
func (ec *ExecutionContext) GetSSHCommander() provision.SSHCommander {
	return ec.sshCommander
}

// GetSSHCommander returns a ssh commander to execute ssh commands against the Minishift VM
func (ec *ExecutionContext) GetOcCommander() *oc.OcRunner {
	return ec.ocRunner
}

// GetDockerCommander returns a commander to run docker commands against the docker daemon used by Minishift
func (ec *ExecutionContext) GetDockerCommander() docker.DockerCommander {
	return ec.dockerCommander
}

func (ec *ExecutionContext) AddToContext(key string, value string) error {
	return ec.interpolationContext.AddToContext(key, value)
}

func (ec *ExecutionContext) RemoveFromContext(key string) error {
	return ec.interpolationContext.RemoveFromContext(key)
}

func (ec *ExecutionContext) Interpolate(cmd string) string {
	return ec.interpolationContext.Interpolate(cmd)
}
