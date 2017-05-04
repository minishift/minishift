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

package manager

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/minishift/addon"
	"github.com/minishift/minishift/pkg/minishift/addon/command"
	"github.com/minishift/minishift/pkg/minishift/addon/parser"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

// AddOnManager is the central point for all operations around managing addons. An addon
// manager is created for the base directory of a addon collection.
type AddOnManager struct {
	baseDir string
	addOns  map[string]addon.AddOn
}

// NewAddOnManager creates a new addon manager for the specified addon directory.
func NewAddOnManager(baseDir string, configMap map[string]*addon.AddOnConfig) (*AddOnManager, error) {
	if !filehelper.IsDirectory(baseDir) {
		return nil, errors.New(fmt.Sprintf("Unable to create addon manager for non existing directory %s", baseDir))
	}

	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create addon manager for non existing directory %s. ", baseDir)
	}

	detectedAddOns := make(map[string]addon.AddOn)
	p := parser.NewAddOnParser()

	for _, f := range files {
		fullPath := filepath.Join(baseDir, f.Name())
		if !filehelper.IsDirectory(fullPath) {
			continue
		}

		addOn, err := p.Parse(fullPath)
		if err != nil {
			_, ok := err.(parser.ParseError)
			if ok {
				glog.Warning(fmt.Sprintf("Skipping addon '%s' in '%s' due to parse error: %s", f.Name(), fullPath, err.Error()))
				continue
			} else {
				return nil, errors.Wrapf(err, "Unable to create addon manager for %s. ", baseDir)
			}
		}
		setStateAndPriority(addOn, configMap)
		detectedAddOns[addOn.MetaData().Name()] = addOn
	}

	return &AddOnManager{baseDir: baseDir, addOns: detectedAddOns}, nil
}

// BaseDir returns the base directory against which this addon mananager was initialised
func (m *AddOnManager) BaseDir() string {
	return m.baseDir
}

// List provides a list of addons managed by this manager
func (m *AddOnManager) List() []addon.AddOn {
	return m.mapToSlice()
}

// Install installs the addon provided via the source directory into the addon directory managed by this addon manager.
// It returns the name of the installed addon. In case an error occurs the empty string and an error are returned.
func (m *AddOnManager) Install(source string, force bool) (string, error) {
	// For now we are assuming that we are dealing with a file path. This can be extended to support URLs, etc as well (HF)
	if !filehelper.IsDirectory(source) {
		return "", errors.New(fmt.Sprintf("The source of a addon needs to be a directory. '%s' is not", source))
	}

	p := parser.NewAddOnParser()
	addOn, err := p.Parse(source)
	if err != nil {
		return "", errors.Wrap(err, "Unable to parse specified addon")
	}

	targetPath := filepath.Join(m.baseDir, filepath.Base(source))
	if filehelper.IsDirectory(targetPath) {
		if force {
			os.RemoveAll(targetPath)
		} else {
			return "", errors.New(fmt.Sprintf("Addon already exists in target directory '%s'", targetPath))
		}
	}

	filehelper.CopyDir(source, targetPath)
	return addOn.MetaData().Name(), nil
}

// Get returns the addon with the specified name. nil is returned if there is no addon with this name.
func (m *AddOnManager) Get(name string) addon.AddOn {
	return m.addOns[name]
}

// Enable enables the addon specified via addonName to be run during startup. The priority determines when the addon is run in relation
// the other addons.
func (m *AddOnManager) Enable(addonName string, priority int) (*addon.AddOnConfig, error) {
	addOn := m.addOns[addonName]
	if addOn == nil {
		return nil, errors.New(fmt.Sprintf("Unable to find addon %s in addon directory %s", addonName, m.baseDir))
	}

	addOn.SetEnabled(true)
	addOn.SetPriority(priority)

	return &addon.AddOnConfig{addonName, true, float64(priority)}, nil
}

// Disable disables the addon with the specified name.
func (m *AddOnManager) Disable(addonName string) (*addon.AddOnConfig, error) {
	addOn := m.addOns[addonName]
	if addOn == nil {
		return nil, errors.New(fmt.Sprintf("Unable to find addon %s in addon directory %s", addonName, m.baseDir))
	}

	addOn.SetEnabled(false)

	return &addon.AddOnConfig{addonName, false, float64(addOn.GetPriority())}, nil
}

// Apply executes all enabled addons.
func (m *AddOnManager) Apply(context *command.ExecutionContext) error {
	addOns := m.mapToSlice()
	sort.Sort(addon.ByPriority(addOns))

	for _, addOn := range addOns {
		if addOn.IsEnabled() {
			err := m.applyAddOn(addOn, context)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *AddOnManager) IsInstalled(name string) bool {
	_, installed := m.addOns[name]
	return installed
}

func (m *AddOnManager) String() string {
	return fmt.Sprintf("%#v", m)
}

func (m *AddOnManager) applyAddOn(addOn addon.AddOn, context *command.ExecutionContext) error {
	fmt.Print(fmt.Sprintf("-- Applying addon '%s':", addOn.MetaData().Name()))
	context.AddToContext("addon-name", addOn.MetaData().Name())
	defer context.RemoveFromContext("addon-name")

	oldDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Unable to apply addon due to failing IO operation")
	}
	defer os.Chdir(oldDir)

	os.Chdir(addOn.InstallPath())
	for _, c := range addOn.Commands() {
		err := c.Execute(context)
		if err != nil {
			return err
		}
	}
	fmt.Print("\n\n")
	return nil
}

func (m *AddOnManager) mapToSlice() []addon.AddOn {
	addOnSlice := make([]addon.AddOn, len(m.addOns))
	i := 0
	for _, value := range m.addOns {
		addOnSlice[i] = value
		i++
	}
	return addOnSlice
}

func setStateAndPriority(addOn addon.AddOn, configMap map[string]*addon.AddOnConfig) {
	addOnConfig := configMap[addOn.MetaData().Name()]
	if addOnConfig == nil {
		return
	}
	addOn.SetEnabled(addOnConfig.Enabled)
	addOn.SetPriority(int(addOnConfig.Priority))
}
