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

package hostfolder

import (
	"fmt"
	"runtime"

	"errors"
	miniConfig "github.com/minishift/minishift/pkg/minishift/config"
	hostFolderConfig "github.com/minishift/minishift/pkg/minishift/hostfolder"
	"github.com/minishift/minishift/pkg/minishift/hostfolder/config"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/minishift/minishift/pkg/util/strings"
	"github.com/spf13/cobra"
)

const (
	usage                = "Usage: minishift hostfolder add --type TYPE --source SOURCE --target TARGET HOST_FOLDER_NAME"
	interactiveModeUsage = "Usage: minishift hostfolder add --interactive --type TYPE [sshfs|cifs (Default: cifs)] HOST_FOLDER_NAME"
	noSource             = "you need to specify the source of the host folder"
	noTarget             = "you need to specify the target of the host folder"
	noUserName           = "you need to specify a username"
	noPassword           = "you need to specify a password"
	noDomain             = "you need to specify the Windows domain"
	unknownType          = "'%s' is an unknown host folder type"
	nonSupportedTtyError = "not a tty supported terminal"
	shareTypeFlag        = "type"
	sourceFlag           = "source"
	targetFlag           = "target"
	optionsFlag          = "options"
	interactiveFlag      = "interactive"
	instanceOnlyFlag     = "instance-only"
	usersShareFlag       = "users-share"
)

var (
	instanceOnly bool
	usersShare   bool
	interactive  bool
	shareType    string
	source       string
	target       string
	options      string
)

var addCmd = &cobra.Command{
	Use:   "add HOST_FOLDER_NAME",
	Short: "Adds a host folder config.",
	Long:  `Adds a host folder config. The defined host folder can be mounted into the Minishift VM file system.`,
	Run:   addHostFolder,
}

func init() {
	HostFolderCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&shareType, shareTypeFlag, "t", "cifs", "The host folder type. Allowed types are [cifs|sshfs]. SSHFS is experimental.")
	addCmd.Flags().StringVar(&source, sourceFlag, "", "The source of the host folder.")
	addCmd.Flags().StringVar(&target, targetFlag, "", "The target (mount point) of the host folder.")
	addCmd.Flags().StringVar(&options, optionsFlag, "", "Host folder type specific options.")
	addCmd.Flags().BoolVar(&instanceOnly, instanceOnlyFlag, false, "Defines the host folder only for the current Minishift instance.")
	addCmd.Flags().BoolVarP(&interactive, interactiveFlag, "i", false, "Allows to interactively provide the required parameters.")

	// Windows-only
	if runtime.GOOS == "windows" {
		addCmd.Flags().BoolVar(&usersShare, usersShareFlag, false, "Defines the shared Users folder as the host folder on a Windows host.")
	}
}

func addHostFolder(cmd *cobra.Command, args []string) {
	if interactive && !util.IsTtySupported() {
		atexit.ExitWithMessage(1, nonSupportedTtyError)
	}

	hostFolderManager := getHostFolderManager()

	var name string
	if usersShare && runtime.GOOS == "windows" {
		// Windows-only (CIFS), all instances
		name = "Users"
	} else {
		if interactive && len(args) < 1 {
			atexit.ExitWithMessage(1, interactiveModeUsage)
		}
		if len(args) < 1 {
			atexit.ExitWithMessage(1, usage)
		}
		name = args[0]
	}

	if hostFolderManager.Exist(name) {
		atexit.ExitWithMessage(1, fmt.Sprintf("there is already a host folder with the name '%s' defined", name))
	}

	switch shareType {
	case hostFolderConfig.CIFS.String():
		if interactive {
			addCIFSInteractive(hostFolderManager, name)
		} else {
			addCIFSNonInteractive(hostFolderManager, name)
		}
	case hostFolderConfig.SSHFS.String():
		if interactive {
			addSSHFSInteractive(hostFolderManager, name)
		} else {
			addSSHFSNonInteractive(hostFolderManager, name)
		}
	default:
		atexit.ExitWithMessage(1, fmt.Sprintf(unknownType, shareType))
	}
}

func addSSHFSInteractive(manager *hostFolderConfig.Manager, name string) error {
	source := util.ReadInputFromStdin("source path")
	if runtime.GOOS == "windows" {
		source = strings.ConvertSlashes(source)
	}

	mountPath := readInputForMountPoint(name)

	config := config.HostFolderConfig{
		Name: name,
		Type: hostFolderConfig.SSHFS.String(),
		Options: map[string]string{
			config.Source:     source,
			config.MountPoint: mountPath,
		},
	}
	hostFolder := hostFolderConfig.NewSSHFSHostFolder(config, miniConfig.AllInstancesConfig)
	manager.Add(hostFolder, !instanceOnly)

	return nil
}

func addSSHFSNonInteractive(manager *hostFolderConfig.Manager, name string) error {
	if source == "" {
		atexit.ExitWithMessage(1, noSource)
	}

	if runtime.GOOS == "windows" {
		source = strings.ConvertSlashes(source)
	}

	if target == "" {
		atexit.ExitWithMessage(1, noTarget)
	}

	config := config.HostFolderConfig{
		Name: name,
		Type: hostFolderConfig.SSHFS.String(),
		Options: map[string]string{
			config.Source:     source,
			config.MountPoint: target,
		},
	}
	hostFolder := hostFolderConfig.NewSSHFSHostFolder(config, miniConfig.AllInstancesConfig)
	manager.Add(hostFolder, !instanceOnly)

	return nil
}

func addCIFSInteractive(manager *hostFolderConfig.Manager, name string) error {
	var uncPath string
	if usersShare {
		uncPath = "[determined on startup]"
	} else {
		uncPath = util.ReadInputFromStdin("UNC path")
	}

	if len(uncPath) == 0 {
		return errors.New("no remote path has been specified")
	}

	mountPoint := readInputForMountPoint(name)

	username := util.ReadInputFromStdin("Username")

	password := util.ReadPasswordFromStdin("Password")
	password, err := util.EncryptText(password)
	if err != nil {
		return err
	}

	domain := util.ReadInputFromStdin("Domain")

	config := config.HostFolderConfig{
		Name: name,
		Type: hostFolderConfig.CIFS.String(),
		Options: map[string]string{
			config.MountPoint: mountPoint,
			config.UncPath:    strings.ConvertSlashes(uncPath),
			config.UserName:   username,
			config.Password:   password,
			config.Domain:     domain,
		},
	}
	hostFolder := hostFolderConfig.NewCifsHostFolder(config)
	manager.Add(hostFolder, !instanceOnly)

	return nil
}

func addCIFSNonInteractive(manager *hostFolderConfig.Manager, name string) {
	if source == "" {
		atexit.ExitWithMessage(1, noSource)
	}

	if target == "" {
		atexit.ExitWithMessage(1, noTarget)
	}

	optionsMap := getOptions(options)

	if _, ok := optionsMap[config.UserName]; !ok {
		atexit.ExitWithMessage(1, noUserName)
	}

	if _, ok := optionsMap[config.Password]; !ok {
		atexit.ExitWithMessage(1, noPassword)
	}

	var domain string
	var ok bool
	if domain, ok = optionsMap[config.Domain]; !ok {
		domain = ""
	}

	config := config.HostFolderConfig{
		Name: name,
		Type: hostFolderConfig.CIFS.String(),
		Options: map[string]string{
			config.UncPath:    strings.ConvertSlashes(source),
			config.MountPoint: target,
			config.UserName:   optionsMap[config.UserName],
			config.Password:   optionsMap[config.Password],
			config.Domain:     domain,
		},
	}
	hostFolder := hostFolderConfig.NewCifsHostFolder(config)
	manager.Add(hostFolder, !instanceOnly)
}
