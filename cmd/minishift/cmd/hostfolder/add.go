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
	"strings"

	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
	hostFolderConfig "github.com/minishift/minishift/pkg/minishift/hostfolder"
	"github.com/minishift/minishift/pkg/minishift/hostfolder/config"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	minishiftStrings "github.com/minishift/minishift/pkg/util/strings"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const (
	usage                = "Usage: minishift hostfolder add --type TYPE --source SOURCE --target TARGET HOST_FOLDER_NAME"
	noName               = "you need to specify a name"
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
	addCmd.Flags().StringVarP(&shareType, shareTypeFlag, "t", "sshfs", "The host folder type. Allowed types are [cifs|sshfs].")
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
		if len(args) < 1 && !interactive {
			atexit.ExitWithMessage(1, usage)
		}
		if interactive {
			name, shareType = readNameAndTypeInteractive()
		} else {
			name = strings.TrimSpace(args[0])
			if len(name) == 0 {
				atexit.ExitWithMessage(1, usage)
			}
		}
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
	source := util.ReadInputFromStdin("Source path")
	source, err := homedir.Expand(source)
	if err != nil {
		atexit.ExitWithMessage(1, err.Error())
	}

	if source == "" {
		atexit.ExitWithMessage(1, noSource)
	}

	if runtime.GOOS == "windows" {
		source = minishiftStrings.ConvertSlashes(source)
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
	hostFolder := hostFolderConfig.NewSSHFSHostFolder(config, minishiftConfig.AllInstancesConfig)
	manager.Add(hostFolder, !instanceOnly)

	return nil
}

func addSSHFSNonInteractive(manager *hostFolderConfig.Manager, name string) error {
	if source == "" {
		atexit.ExitWithMessage(1, noSource)
	}

	if runtime.GOOS == "windows" {
		source = minishiftStrings.ConvertSlashes(source)
	}

	if target == "" {
		atexit.ExitWithMessage(1, noTarget)
	}

	config := config.HostFolderConfig{
		Name: name,
		Type: hostFolderConfig.SSHFS.String(),
		Options: map[string]string{
			config.Source:       source,
			config.MountPoint:   target,
			config.ExtraOptions: options,
		},
	}
	hostFolder := hostFolderConfig.NewSSHFSHostFolder(config, minishiftConfig.AllInstancesConfig)
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

	if uncPath == "" {
		atexit.ExitWithMessage(1, noSource)
	}

	mountPoint := readInputForMountPoint(name)

	username := util.ReadInputFromStdin("Username")
	if username == "" {
		atexit.ExitWithMessage(1, noUserName)
	}

	password := util.ReadPasswordFromStdin("Password")
	if password == "" {
		atexit.ExitWithMessage(1, noPassword)
	}

	password, err := util.EncryptText(password)
	if err != nil {
		return err
	}

	domain := util.ReadInputFromStdin("Domain")

	config := config.HostFolderConfig{
		Name: name,
		Type: hostFolderConfig.CIFS.String(),
		Options: map[string]string{
			config.MountPoint:   mountPoint,
			config.UncPath:      minishiftStrings.ConvertSlashes(uncPath),
			config.UserName:     username,
			config.Password:     password,
			config.Domain:       domain,
			config.ExtraOptions: options,
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
			config.UncPath:    minishiftStrings.ConvertSlashes(source),
			config.MountPoint: target,
			config.UserName:   optionsMap[config.UserName],
			config.Password:   optionsMap[config.Password],
			config.Domain:     domain,
		},
	}
	hostFolder := hostFolderConfig.NewCifsHostFolder(config)
	manager.Add(hostFolder, !instanceOnly)
}

// returns the name and full-name for shareType
func readNameAndTypeInteractive() (string, string) {
	name := util.ReadInputFromStdin("Name")
	if name == "" {
		atexit.ExitWithMessage(1, noName)
	}
	shareType := strings.ToLower(util.ReadInputFromStdin("Type [sshfs, cifs (S/c)]"))

	if shareType == "s" || shareType == "" {
		return name, hostFolderConfig.SSHFS.String()
	}

	if shareType == "c" {
		return name, hostFolderConfig.CIFS.String()
	}

	return name, shareType
}
