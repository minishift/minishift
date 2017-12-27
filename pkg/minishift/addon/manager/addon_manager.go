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
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"strings"

	minikubeConstants "github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/addon"
	"github.com/minishift/minishift/pkg/minishift/addon/command"
	"github.com/minishift/minishift/pkg/minishift/addon/parser"
	"github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/filehelper"
	utilStrings "github.com/minishift/minishift/pkg/util/strings"
	"github.com/pkg/errors"
)

const versionRangeSeparator = ","

// AddOnManager is the central point for all operations around managing addons. An addon
// manager is created for the base directory of a addon collection.
type AddOnManager struct {
	baseDir string
	addOns  map[string]addon.AddOn
}

// NewAddOnManager creates a new addon manager for the specified addon directory.
func NewAddOnManager(baseDir string, configMap map[string]*addon.AddOnConfig) (*AddOnManager, error) {
	if !filehelper.IsDirectory(baseDir) {
		return nil, errors.New(fmt.Sprintf("Unable to create addon manager for non existing directory '%s'", baseDir))
	}

	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create addon manager for non existing directory '%s'. ", baseDir)
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
				fmt.Println(fmt.Sprintf("Skipping addon '%s' in '%s' due to parse error: %s", f.Name(), fullPath, err.Error()))
				continue
			} else {
				return nil, errors.Wrapf(err, "Unable to create addon manager for '%s'. ", baseDir)
			}
		}
		setStateAndPriority(addOn, configMap)
		detectedAddOns[addOn.MetaData().Name()] = addOn
	}

	return &AddOnManager{baseDir: baseDir, addOns: detectedAddOns}, nil
}

// BaseDir returns the base directory against which this addon manager was initialised
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
		return nil, errors.New(fmt.Sprintf("Unable to find addon '%s' in addon directory '%s'", addonName, m.baseDir))
	}

	addOn.SetEnabled(true)
	addOn.SetPriority(priority)

	return &addon.AddOnConfig{addonName, true, float64(priority)}, nil
}

// UnInstall uninstalls the addon specified via addonName.
func (m *AddOnManager) UnInstall(addonName string) error {
	addOn := m.addOns[addonName]
	if addOn == nil {
		return errors.New(fmt.Sprintf("Unable to find addon '%s' in addon directory '%s'", addonName, m.baseDir))
	}

	targetPath := filepath.Join(m.baseDir, addonName)
	if filehelper.IsDirectory(targetPath) {
		if err := os.RemoveAll(targetPath); err != nil {
			return err
		}
		return nil
	}
	return errors.New(fmt.Sprintf("Unable to find addon directory '%s'", targetPath))
}

// Disable disables the addon with the specified name.
func (m *AddOnManager) Disable(addonName string) (*addon.AddOnConfig, error) {
	addOn := m.addOns[addonName]
	if addOn == nil {
		return nil, errors.New(fmt.Sprintf("Unable to find addon '%s' in addon directory '%s'", addonName, m.baseDir))
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
			err := m.ApplyAddOn(addOn, context)
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

func (m *AddOnManager) ApplyAddOn(addOn addon.AddOn, context *command.ExecutionContext) error {
	fmt.Print(fmt.Sprintf("-- Applying addon '%s':", addOn.MetaData().Name()))
	context.AddToContext("addon-name", addOn.MetaData().Name())
	defer context.RemoveFromContext("addon-name")

	if err := addVarDefaultsToContext(addOn, context); err != nil {
		return err
	}

	addonMetadata := addOn.MetaData()
	if err := verifyRequiredOpenshiftVersion(context, addonMetadata); err != nil {
		return err
	}
	if err := verifyRequiredVariablesInContext(context, addonMetadata); err != nil {
		return err
	}

	oldDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Unable to apply addon due to failing IO operation")
	}
	defer os.Chdir(oldDir)

	os.Chdir(addOn.InstallPath())
	if err := addonCmdExecution(addOn.Commands(), context); err != nil {
		return err
	}

	fmt.Print("\n")
	return nil
}

func (m *AddOnManager) RemoveAddOn(addOn addon.AddOn, context *command.ExecutionContext) error {
	fmt.Print(fmt.Sprintf("-- Removing addon '%s':", addOn.MetaData().Name()))
	context.AddToContext("addon-name", addOn.MetaData().Name())
	defer context.RemoveFromContext("addon-name")

	if err := addVarDefaultsToContext(addOn, context); err != nil {
		return err
	}

	addonMetadata := addOn.MetaData()
	if err := verifyRequiredOpenshiftVersion(context, addonMetadata); err != nil {
		return err
	}
	if err := verifyRequiredVariablesInContext(context, addonMetadata); err != nil {
		return err
	}

	oldDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Unable to remove addon due to failing IO operation")
	}
	defer os.Chdir(oldDir)

	os.Chdir(addOn.InstallPath())
	if err := addonCmdExecution(addOn.RemoveCommands(), context); err != nil {
		return err
	}
	fmt.Print("\n")
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

func verifyRequiredVariablesInContext(context *command.ExecutionContext, meta addon.AddOnMeta) error {
	missingVars := []string{}

	check := make(map[string]bool)
	for _, v := range context.Vars() {
		check[v] = true
	}

	requiredVars, err := meta.RequiredVars()
	if err != nil {
		return err
	}
	for _, requiredVar := range requiredVars {
		if !check[requiredVar] {
			missingVars = append(missingVars, requiredVar)
		}
	}

	if len(missingVars) > 0 {
		missing := strings.TrimSpace(strings.Join(missingVars, ", "))
		return fmt.Errorf("The variable(s) '%s' are required by the add-on, but are not defined in the context", missing)
	}

	return nil
}

func verifyRequiredOpenshiftVersion(context *command.ExecutionContext, meta addon.AddOnMeta) error {
	dockerCommander := context.GetDockerCommander()
	versionInfo, err := dockerCommander.Exec(" ", constants.OpenshiftContainerName, "openshift", "version")
	if err != nil {
		return err
	}

	// verionInfo variable have below string as value along with new line
	// openshift v3.6.0+c4dd4cf
	// kubernetes v1.6.1+5115d708d7
	// etcd 3.2.1
	// openShiftVersionAlongWithCommitSha is contain *v3.6.0+c4dd4cf* (first split on new line and second on space)
	openShiftVersionAlongWithCommitSha := strings.Split(strings.Split(versionInfo, "\n")[0], " ")[1]
	// openshiftVersion is contain *3.6.0* (split on *+* string and then trim the *v* as perfix)
	// TrimSpace is there to make sure no whitespace around version string
	openShiftVersion := strings.TrimSpace(strings.TrimPrefix(strings.Split(openShiftVersionAlongWithCommitSha, "+")[0], minikubeConstants.VersionPrefix))
	requiredOpenshiftVersions := strings.TrimSpace(meta.OpenShiftVersion())
	if requiredOpenshiftVersions != "" {
		for _, requiredOpenshiftVersion := range strings.Split(requiredOpenshiftVersions, versionRangeSeparator) {
			if err = compareOpenshiftVersions(openShiftVersion, strings.TrimSpace(requiredOpenshiftVersion)); err != nil {
				return err
			}
		}
	}

	return nil
}

func setStateAndPriority(addOn addon.AddOn, configMap map[string]*addon.AddOnConfig) {
	addOnConfig := configMap[addOn.MetaData().Name()]
	if addOnConfig == nil {
		return
	}
	addOn.SetEnabled(addOnConfig.Enabled)
	addOn.SetPriority(int(addOnConfig.Priority))
}

func compareOpenshiftVersions(openShiftVersion, requiredOpenshiftVersion string) error {
	if strings.HasPrefix(requiredOpenshiftVersion, ">=") {
		// This will work for both upstream and downstream.
		if util.VersionOrdinal(openShiftVersion) < util.VersionOrdinal(strings.TrimPrefix(requiredOpenshiftVersion, ">=")) {
			return fmt.Errorf("\nAdd-on does not support OpenShift version %s. "+
				"You need to use a version %s", openShiftVersion, requiredOpenshiftVersion)
		}
		return nil
	}
	if strings.HasPrefix(requiredOpenshiftVersion, ">") {
		if util.VersionOrdinal(openShiftVersion) <= util.VersionOrdinal(strings.TrimPrefix(requiredOpenshiftVersion, ">")) {
			return fmt.Errorf("\nAdd-on does not support OpenShift version %s. "+
				"You need to use a version %s", openShiftVersion, requiredOpenshiftVersion)
		}
		return nil
	}
	if strings.HasPrefix(requiredOpenshiftVersion, "<=") {
		if util.VersionOrdinal(openShiftVersion) > util.VersionOrdinal(strings.TrimPrefix(requiredOpenshiftVersion, "<=")) {
			return fmt.Errorf("\nAdd-on does not support OpenShift version %s. "+
				"You need to use a version %s", openShiftVersion, requiredOpenshiftVersion)
		}
		return nil
	}
	if strings.HasPrefix(requiredOpenshiftVersion, "<") {
		if util.VersionOrdinal(openShiftVersion) >= util.VersionOrdinal(strings.TrimPrefix(requiredOpenshiftVersion, "<")) {
			return fmt.Errorf("\nAdd-on does not support OpenShift version %s. "+
				"You need to use a version %s", openShiftVersion, requiredOpenshiftVersion)
		}
		return nil
	}
	if openShiftVersion != requiredOpenshiftVersion {
		return fmt.Errorf("\nAddon does not support OpenShift version %s. "+
			"You need to use a version %s", openShiftVersion, requiredOpenshiftVersion)
	}

	return nil
}

func addVarDefaultsToContext(addOn addon.AddOn, context *command.ExecutionContext) error {
	varDefaults, err := addOn.MetaData().VarDefaults()
	if err != nil {
		return err
	}

	for _, varDefault := range varDefaults {
		// Don't add context if env already present
		if !utilStrings.Contains(context.Vars(), varDefault.Key) {
			context.AddToContext(varDefault.Key, varDefault.Value)
		}
	}

	return nil
}

func addonCmdExecution(commands []command.Command, context *command.ExecutionContext) error {
	for _, c := range commands {
		if err := c.Execute(context); err != nil {
			return err
		}
	}

	return nil
}
