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
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/spf13/viper"

	config "github.com/minishift/minishift/pkg/minishift/config"
	miniutil "github.com/minishift/minishift/pkg/minishift/util"
	"github.com/minishift/minishift/pkg/util"
)

const (
	HostfoldersAutoMountKey = "hostfolders-automount"
)

func IsAutoMount() bool {
	return viper.GetBool(HostfoldersAutoMountKey)
}

func isHostRunning(driver drivers.Driver) bool {
	return drivers.MachineInState(driver, state.Running)()
}

func IsHostfoldersDefined() bool {
	return len(config.InstanceConfig.HostFolders) > 0 ||
		len(config.AllInstancesConfig.HostFolders) > 0
}

func isHostfolderDefinedByName(name string) bool {
	return getHostfolderByName(name) != nil
}

func List(driver drivers.Driver, isRunning bool) error {
	if !IsHostfoldersDefined() {
		return errors.New("No host folders defined")
	}

	procMounts := ""
	if isRunning {
		cmd := fmt.Sprintf("cat /proc/mounts")
		procMounts, _ = drivers.RunSSHCommandFromDriver(driver, cmd)
	}

	w := tabwriter.NewWriter(os.Stdout, 4, 8, 3, ' ', 0)
	fmt.Fprintln(w, "Name\tMountpoint\tRemote path\tMounted")

	hostfolders := config.AllInstancesConfig.HostFolders
	hostfolders = append(hostfolders, config.InstanceConfig.HostFolders...)
	for i := range hostfolders {
		hostfolder := hostfolders[i]

		remotePath := ""
		switch hostfolder.Type {
		case "cifs":
			remotePath = hostfolder.Options["uncpath"]
		}

		mounted := "N"
		if isRunning && strings.Contains(procMounts, hostfolder.Mountpoint()) {
			mounted = "Y"
		}

		fmt.Fprintln(w,
			(fmt.Sprintf("%s\t%s\t%s\t%s",
				hostfolder.Name,
				hostfolder.Mountpoint(),
				remotePath,
				mounted)))
	}

	w.Flush()
	return nil
}

func readInputForMountpoint(name string) string {
	defaultMountpoint := config.GetHostfoldersMountPath(name)
	mountpointText := fmt.Sprintf("Mountpoint [%s]", defaultMountpoint)
	return util.ReadInputFromStdin(mountpointText)
}

func SetupUsers(allInstances bool) error {
	name := "Users"
	if isHostfolderDefinedByName(name) {
		return fmt.Errorf("Already have a host folder defined for: %s", name)
	}

	mountpoint := readInputForMountpoint(name)
	username := util.ReadInputFromStdin("Username")
	password := util.ReadPasswordFromStdin("Password")
	domain := util.ReadInputFromStdin("Domain")
	password, err := util.EncryptText(password)
	if err != nil {
		return err
	}

	// We only store this record for credentials purpose
	addToConfig(newCifsHostFolder(
		name,
		"[determined on startup]",
		mountpoint,
		username, password, domain),
		allInstances)

	return nil
}

func Add(name string, allInstances bool) error {
	if isHostfolderDefinedByName(name) {
		return fmt.Errorf("Already have a host folder defined for: %s", name)
	}

	uncpath := util.ReadInputFromStdin("UNC path")
	if len(uncpath) == 0 {
		return fmt.Errorf("No remote path has been given")
	}
	mountpoint := readInputForMountpoint(name)
	username := util.ReadInputFromStdin("Username")
	password := util.ReadPasswordFromStdin("Password")
	domain := util.ReadInputFromStdin("Domain")
	password, err := util.EncryptText(password)
	if err != nil {
		return err
	}

	addToConfig(newCifsHostFolder(
		name,
		uncpath,
		mountpoint,
		username, password, domain),
		allInstances)

	return nil
}

func newCifsHostFolder(name string, uncpath string, mountpoint string, username string, password string, domain string) config.HostFolder {
	return config.HostFolder{
		Name: name,
		Type: "cifs",
		Options: map[string]string{
			"mountpoint": mountpoint,
			"uncpath":    convertSlashes(uncpath),
			"username":   username,
			"password":   password,
			"domain":     domain,
		},
	}
}

func addToConfig(hostfolder config.HostFolder, allInstances bool) {
	if allInstances {
		config.AllInstancesConfig.HostFolders = append(config.AllInstancesConfig.HostFolders, hostfolder)
		config.AllInstancesConfig.Write()
	} else {
		config.InstanceConfig.HostFolders = append(config.InstanceConfig.HostFolders, hostfolder)
		config.InstanceConfig.Write()
	}

	println(fmt.Sprintf("Added: %s", hostfolder.Name))
}

func Remove(name string) error {
	if !isHostfolderDefinedByName(name) {
		return fmt.Errorf("No host folder defined as: %s", name)
	}

	config.InstanceConfig.HostFolders = removeFromHostFoldersByName(name, config.InstanceConfig.HostFolders)
	config.InstanceConfig.Write()

	config.AllInstancesConfig.HostFolders = removeFromHostFoldersByName(name, config.AllInstancesConfig.HostFolders)
	config.AllInstancesConfig.Write()

	println(fmt.Sprintf("Removed: %s", name))

	return nil
}

func Mount(driver drivers.Driver, name string) error {
	if !isHostRunning(driver) {
		return errors.New("Host is in the wrong state.")
	}

	if !IsHostfoldersDefined() {
		return errors.New("No host folders defined.")
	}

	hostfolder := getHostfolderByName(name)
	if hostfolder == nil {
		return fmt.Errorf("No host folder defined as: %s", name)
	} else {
		ensureMountPointExists(driver, hostfolder)
		mountHostfolder(driver, hostfolder)
	}
	return nil
}

// Performs mounting of host folders
func MountHostfolders(driver drivers.Driver) error {
	if !isHostRunning(driver) {
		return errors.New("Host is in the wrong state.")
	}

	if !IsHostfoldersDefined() {
		return errors.New("No host folders defined.")
	}

	println("-- Mounting hostfolders")

	hostfolders := config.AllInstancesConfig.HostFolders
	hostfolders = append(hostfolders, config.InstanceConfig.HostFolders...)
	for i := range hostfolders {
		mountHostfolder(driver, &hostfolders[i])
	}

	return nil
}

func Umount(driver drivers.Driver, name string) error {
	if !isHostRunning(driver) {
		return errors.New("Host is in the wrong state.")
	}

	if !IsHostfoldersDefined() {
		return errors.New("No host folders defined")
	}

	hostfolder := getHostfolderByName(name)
	if hostfolder == nil {
		return fmt.Errorf("No host folder defined as: %s", name)
	} else {
		umountHostfolder(driver, hostfolder)
	}
	return nil
}

func mountHostfolder(driver drivers.Driver, hostfolder *config.HostFolder) error {
	if hostfolder == nil {
		return errors.New("Host folder not defined")
	}

	switch hostfolder.Type {
	case "cifs":
		if err := mountCifsHostfolder(driver, hostfolder); err != nil {
			return err
		}
	default:
		return errors.New("Unsupported host folder type")
	}

	return nil
}

func determineHostIp(driver drivers.Driver) (string, error) {
	instanceip, err := driver.GetIP()
	if err != nil {
		return "", err
	}

	for _, hostaddr := range miniutil.HostIPs() {

		if miniutil.NetworkContains(hostaddr, instanceip) {
			hostip, _, _ := net.ParseCIDR(hostaddr)
			if miniutil.IsIPReachable(driver, hostip.String(), false) {
				return hostip.String(), nil
			} else {
				return "", errors.New("Unreachable")
			}
		}
	}

	return "", errors.New("Unknown error occured")
}

func mountCifsHostfolder(driver drivers.Driver, hostfolder *config.HostFolder) error {
	// If "Users" is used as name, determine the IP of host for UNC path on startup
	if hostfolder.Name == "Users" {
		hostip, _ := determineHostIp(driver)
		hostfolder.Options["uncpath"] = fmt.Sprintf("//%s/Users", hostip)
	}

	print(fmt.Sprintf("   Mounting '%s': '%s' as '%s' ... ",
		hostfolder.Name,
		hostfolder.Options["uncpath"],
		hostfolder.Mountpoint()))

	if isMounted, err := isHostfolderMounted(driver, hostfolder); isMounted {
		println("Already mounted")
		return fmt.Errorf("Host folder is already mounted. %s", err)
	}

	if !isCifsHostReachable(driver, hostfolder.Options["uncpath"]) {
		print("Unreachable\n")
		return errors.New("Host folder is unreachable")
	}

	password, err := util.DecryptText(hostfolder.Options["password"])
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf(
		"sudo mount -t cifs %s %s -o username=%s,password=%s",
		hostfolder.Options["uncpath"],
		hostfolder.Mountpoint(),
		hostfolder.Options["username"],
		password)

	if len(hostfolder.Options["domain"]) > 0 { // != ""
		cmd = fmt.Sprintf("%s,domain=%s", cmd, hostfolder.Options["domain"])
	}

	if err := ensureMountPointExists(driver, hostfolder); err != nil {
		println("FAIL")
		return fmt.Errorf("Error occured while creating mountpoint. %s", err)
	}

	if _, err := drivers.RunSSHCommandFromDriver(driver, cmd); err != nil {
		println("FAIL")
		return fmt.Errorf("Error occured while mounting host folder.", err)
	} else {
		println("OK")
	}

	return nil
}

func umountHostfolder(driver drivers.Driver, hostfolder *config.HostFolder) error {
	if hostfolder == nil {
		errors.New("Host folder not defined")
	}

	print(fmt.Sprintf("   Unmounting '%s' ... ", hostfolder.Name))

	if isMounted, err := isHostfolderMounted(driver, hostfolder); !isMounted {
		print("Not mounted\n")
		return fmt.Errorf("Host folder not mounted. %s", err)
	}

	cmd := fmt.Sprintf(
		"sudo umount %s",
		hostfolder.Mountpoint())

	if _, err := drivers.RunSSHCommandFromDriver(driver, cmd); err != nil {
		println("FAIL")
		return fmt.Errorf("Error occured while unmounting host folder.", err)
	} else {
		println("OK")
	}

	return nil
}

func isHostfolderMounted(driver drivers.Driver, hostfolder *config.HostFolder) (bool, error) {
	cmd := fmt.Sprintf(
		"if grep -qs %s /proc/mounts; then echo '1'; else echo '0'; fi",
		hostfolder.Mountpoint())

	out, err := drivers.RunSSHCommandFromDriver(driver, cmd)

	if err != nil {
		return false, err
	}
	if strings.Trim(out, "\n") == "0" {
		return false, nil
	}

	return true, nil
}

func convertSlashes(input string) string {
	return strings.Replace(input, "\\", "/", -1)
}

func isCifsHostReachable(driver drivers.Driver, uncpath string) bool {
	host := ""

	splithost := strings.Split(uncpath, "/")
	if len(splithost) > 2 {
		host = splithost[2]
	}

	if host == "" {
		return false
	}

	return miniutil.IsIPReachable(driver, host, false)
}

func ensureMountPointExists(driver drivers.Driver, hostfolder *config.HostFolder) error {
	if hostfolder == nil {
		errors.New("Host folder is not defined")
	}

	cmd := fmt.Sprintf(
		"sudo mkdir -p %s",
		hostfolder.Mountpoint())

	if _, err := drivers.RunSSHCommandFromDriver(driver, cmd); err != nil {
		return err
	}

	return nil
}

func removeFromHostFoldersByName(name string, hostfolders []config.HostFolder) []config.HostFolder {
	for i := range hostfolders {

		hostfolder := hostfolders[i]

		if hostfolder.Name == name {
			hostfolders = append(hostfolders[:i], hostfolders[i+1:]...)
			break
		}
	}
	return hostfolders
}

func getHostfolderByName(name string) *config.HostFolder {
	hostfolder := getHostfolderByNameFromList(name, config.InstanceConfig.HostFolders)
	if hostfolder != nil {
		return hostfolder
	}

	return getHostfolderByNameFromList(name, config.AllInstancesConfig.HostFolders)
}

func getHostfolderByNameFromList(name string, hostfolders []config.HostFolder) *config.HostFolder {
	for i := range hostfolders {

		hostfolder := hostfolders[i]

		if hostfolder.Name == name {
			return &hostfolder
		}
	}

	return nil
}
