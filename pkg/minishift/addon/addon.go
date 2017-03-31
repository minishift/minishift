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

package addon

import (
	"fmt"
	"github.com/minishift/minishift/pkg/minishift/addon/command"
)

// AddOn is the internal representation of an AddOn including the addon metadata and the actual commands
// the Addon consists off.
type AddOn interface {
	// GetMetaData returns the meta data of this addon.
	MetaData() AddOnMeta

	// Commands returns a Command slice of the commands this addon consists of.
	Commands() []command.Command

	// InstallPath returns the path under which the addon is installed.
	InstallPath() string

	// IsEnabled returns true, if this addon is currently enabled and getting applied as part of OpenShift
	// provisioning. Returns false, is this addon is not enabled.
	IsEnabled() bool

	// SetEnabled sets this addon as en- or disabled.
	SetEnabled(enabled bool)

	// GetPriority returns the priority of this add on in terms of the order in which it gets applied
	GetPriority() int

	// SetPriority sets the priority of this addon in terms of its installation order. Addons with lower priority
	// gets installed first
	SetPriority(priority int)
}

// DefaultAddOn provides the default implementation of the AddOn interface
type DefaultAddOn struct {
	AddOn

	metaData AddOnMeta
	commands []command.Command
	enabled  bool
	path     string
	priority int
}

func NewAddOn(meta AddOnMeta, commands []command.Command, path string) AddOn {
	addOn := DefaultAddOn{
		metaData: meta,
		commands: commands,
		path:     path,
	}
	return &addOn
}

func (a *DefaultAddOn) Commands() []command.Command {
	return a.commands
}

func (a *DefaultAddOn) MetaData() AddOnMeta {
	return a.metaData
}

func (a *DefaultAddOn) IsEnabled() bool {
	return a.enabled
}

func (a *DefaultAddOn) SetEnabled(enabled bool) {
	a.enabled = enabled
}

func (a *DefaultAddOn) SetPriority(priority int) {
	a.priority = priority
}

func (a *DefaultAddOn) GetPriority() int {
	return a.priority
}

func (a *DefaultAddOn) InstallPath() string {
	return a.path
}

func (meta *DefaultAddOn) String() string {
	return fmt.Sprintf("%#v", meta)
}

// ByPriority implements sort.Interface for []AddOn based on the Priority field.
type ByPriority []AddOn

func (a ByPriority) Len() int           { return len(a) }
func (a ByPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriority) Less(i, j int) bool { return a[i].GetPriority() < a[j].GetPriority() }

// ByStatusThenName implements sort.Interface for []AddOn based primarily on the addon enable state and secondarily
// on the addon name.
type ByStatusThenName []AddOn

func (a ByStatusThenName) Len() int      { return len(a) }
func (a ByStatusThenName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByStatusThenName) Less(i, j int) bool {
	if a[i].IsEnabled() && !a[j].IsEnabled() {
		return true
	}

	if !a[i].IsEnabled() && a[j].IsEnabled() {
		return false
	}

	return a[i].MetaData().Name() < a[j].MetaData().Name()
}

// ByStatusThenPriority implements sort.Interface for []AddOn based primarily on the addon enable state and secondarily
// on the addon priority. Lastly the name is considered
type ByStatusThenPriorityThenName []AddOn

func (a ByStatusThenPriorityThenName) Len() int      { return len(a) }
func (a ByStatusThenPriorityThenName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByStatusThenPriorityThenName) Less(i, j int) bool {
	if a[i].IsEnabled() && !a[j].IsEnabled() {
		return true
	}

	if !a[i].IsEnabled() && a[j].IsEnabled() {
		return false
	}

	if a[i].GetPriority() < a[j].GetPriority() {
		return true
	}

	if a[i].GetPriority() < a[j].GetPriority() {
		return false
	}

	return a[i].MetaData().Name() < a[j].MetaData().Name()
}
