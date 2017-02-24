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

	instanceState "github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/minishift/util"
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

func IsHostfoldersDefined(printMessage bool) bool {
	if len(instanceState.Config.HostFolders) == 0 {
		if printMessage {
			println("No host folders defined")
		}
		return false
	}
	return true
}

func isHostfolderDefinedByName(name string, printMessage bool) bool {
	hostfolder := getHostfolderByName(name)
	if hostfolder != nil {
		if printMessage {
			println(fmt.Sprintf("Already have a host folder defined for: %s", name))
		}
		return true
	}
	return false
}

func List(driver drivers.Driver) {
	if !IsHostfoldersDefined(true) {
		return
	}

	procMounts := ""
	isRunning := isHostRunning(driver)
	if isRunning {
		cmd := fmt.Sprintf("cat /proc/mounts")
		procMounts, _ = drivers.RunSSHCommandFromDriver(driver, cmd)
	}

	w := tabwriter.NewWriter(os.Stdout, 4, 8, 3, ' ', 0)
	fmt.Fprintln(w, "Name\tMountpoint\tRemote path\tMounted")

	hostfolders := instanceState.Config.HostFolders
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
}

func readInputForMountpoint(name string) string {
	defaultMountpoint := instanceState.GetHostfoldersMountPath(name)
	mountpointText := fmt.Sprintf("Mountpoint [%s]", defaultMountpoint)
	return util.ReadInputFromStdin(mountpointText)
}

func SetupUsers() {
	name := "Users"
	if !isHostfolderDefinedByName(name, true) {
		mountpoint := readInputForMountpoint(name)
		username := util.ReadInputFromStdin("Username")
		password := util.ReadPasswordFromStdin("Password")
		domain := util.ReadInputFromStdin("Domain")

		// We only store this record for credentials purpose
		addToConfig(newCifsHostFolder(
			name,
			"[determined on startup]",
			mountpoint,
			username, password, domain))
	}
}

func Add(name string) {
	if !isHostfolderDefinedByName(name, true) {
		mountpoint := readInputForMountpoint(name)
		uncpath := util.ReadInputFromStdin("UNC path")
		username := util.ReadInputFromStdin("Username")
		password := util.ReadPasswordFromStdin("Password")
		domain := util.ReadInputFromStdin("Domain")

		addToConfig(newCifsHostFolder(
			name,
			uncpath,
			mountpoint,
			username, password, domain))
	}
}

func newCifsHostFolder(name string, uncpath string, mountpoint string, username string, password string, domain string) instanceState.HostFolder {
	return instanceState.HostFolder{
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

func addToConfig(hostfolder instanceState.HostFolder) {
	instanceState.Config.HostFolders = append(instanceState.Config.HostFolders, hostfolder)
	instanceState.Config.Write()

	println(fmt.Sprintf("Added: %s", hostfolder.Name))
}

func Remove(name string) {
	if isHostfolderDefinedByName(name, false) {
		hostfolders := instanceState.Config.HostFolders

		for i := range hostfolders {

			hostfolder := hostfolders[i]

			if hostfolder.Name == name {
				hostfolders = append(hostfolders[:i], hostfolders[i+1:]...)
				break
			}
		}

		instanceState.Config.HostFolders = hostfolders
		instanceState.Config.Write()
		println(fmt.Sprintf("Removed: %s", name))
	}
}

func Mount(driver drivers.Driver, name string) error {
	if !isHostRunning(driver) {
		return errors.New("Host is in the wrong state.")
	}

	if !IsHostfoldersDefined(true) {
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

	if !IsHostfoldersDefined(true) {
		return errors.New("No host folders defined.")
	}

	println("-- Mounting hostfolders")

	hostfolders := instanceState.Config.HostFolders
	for i := range hostfolders {
		// Ignore individual errors or aggregate
		mountHostfolder(driver, &hostfolders[i])
	}

	return nil
}

func Umount(driver drivers.Driver, name string) error {
	if !isHostRunning(driver) {
		return errors.New("Host is in the wrong state.")
	}

	if !IsHostfoldersDefined(true) {
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

func mountHostfolder(driver drivers.Driver, hostfolder *instanceState.HostFolder) error {
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

	for _, hostaddr := range util.HostIPs() {

		if util.NetworkContains(hostaddr, instanceip) {
			hostip, _, _ := net.ParseCIDR(hostaddr)
			if util.IsIPReachable(driver, hostip.String(), false) {
				return hostip.String(), nil
			} else {
				return "", errors.New("Unreachable")
			}
		}
	}

	return "", errors.New("Unknown error occured")
}

func mountCifsHostfolder(driver drivers.Driver, hostfolder *instanceState.HostFolder) error {
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

	cmd := fmt.Sprintf(
		"sudo mount -t cifs %s %s -o username=%s,password=%s",
		hostfolder.Options["uncpath"],
		hostfolder.Mountpoint(),
		hostfolder.Options["username"],
		hostfolder.Options["password"])

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

func umountHostfolder(driver drivers.Driver, hostfolder *instanceState.HostFolder) error {
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

func isHostfolderMounted(driver drivers.Driver, hostfolder *instanceState.HostFolder) (bool, error) {
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

	return util.IsIPReachable(driver, host, false)
}

func ensureMountPointExists(driver drivers.Driver, hostfolder *instanceState.HostFolder) error {
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

func getHostfolderByName(name string) *instanceState.HostFolder {
	hostfolders := instanceState.Config.HostFolders
	for i := range hostfolders {

		hostfolder := hostfolders[i]

		if hostfolder.Name == name {
			return &hostfolder
		}
	}

	return nil
}
