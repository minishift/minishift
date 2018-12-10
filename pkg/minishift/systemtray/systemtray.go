// +build !systemtray

/*
Copyright (C) 2018 Red Hat, Inc.

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

package systemtray

import (
	"fmt"
	goos "os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/anjannath/systray"
	"github.com/docker/machine/libmachine"
	"github.com/golang/glog"
	cmdUtil "github.com/minishift/minishift/cmd/minishift/cmd/util"
	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/profile"
	"github.com/minishift/minishift/pkg/minishift/shell/powershell"
	"github.com/minishift/minishift/pkg/minishift/systemtray/icon"
	"github.com/minishift/minishift/pkg/util/os"
	"github.com/pkg/browser"
)

const (
	WEB_CONSOLE string = "Web Console"
	START       string = "Start"
	STOP        string = "Stop"
	EXIT        string = "Exit"
)

const (
	DOES_NOT_EXIST int = iota
	RUNNING
	STOPPED
	START_PROFILE
	STOP_PROFILE
)

var (
	submenus            = make(map[string]*systray.MenuItem)
	submenusToMenuItems = make(map[string]MenuAction)

	profiles        []string
	profileMenuList []*systray.MenuItem

	submenusLock            sync.RWMutex
	submenusToMenuItemsLock sync.RWMutex
)

type MenuAction struct {
	start   *systray.MenuItem
	stop    *systray.MenuItem
	console *systray.MenuItem
}

func OnReady() {
	systray.SetIcon(icon.TrayIcon)
	exit := systray.AddMenuItem(EXIT, "", 0)
	systray.AddSeparator()
	profiles = profile.GetProfileList()
	for _, profile := range profiles {
		submenu := systray.AddSubMenu(strings.Title(profile))
		startMenu := submenu.AddSubMenuItem(START, "", 0)
		stopMenu := submenu.AddSubMenuItem(STOP, "", 0)
		consoleMenu := submenu.AddSubMenuItem(WEB_CONSOLE, "", 0)
		submenus[profile] = submenu
		submenusToMenuItems[profile] = MenuAction{start: startMenu, stop: stopMenu, console: consoleMenu}
	}

	go func() {
		<-exit.OnClickCh()
		systray.Quit()
	}()

	for k, v := range submenusToMenuItems {
		go startStopHandler(icon.Running, k, v.start, START_PROFILE)
		go startStopHandler(icon.Stopped, k, v.stop, STOP_PROFILE)
		go webConsoleHandler(k, v.console)
	}

	go updateTrayMenu()

	go updateProfileStatus()
}

func OnExit() {
	return
}

func getStatus(profileName string) int {
	cmd, _ := os.CurrentExecutable()
	args := []string{"status", "--profile", profileName}
	command := exec.Command(cmd, args...)
	out, _ := command.Output()
	stdOut := fmt.Sprintf("%s", out)
	fmt.Println(stdOut)

	if strings.Contains(stdOut, "Running") {
		return RUNNING
	} else {
		return STOPPED
	}
}

// Add newly created profiles and remove deleted profiles from tray
func updateTrayMenu() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Failed to create watcher:", err)
	}
	profilesBaseDir := constants.GetMinishiftProfilesDir()
	err = watcher.Add(profilesBaseDir)
	if err != nil {
		fmt.Println("Failed to watch profiles directory:", err)
	}

	for {
		event, _ := <-watcher.Events
		if event.Op&fsnotify.Remove == fsnotify.Remove {
			profile := filepath.Base(event.Name)
			if _, ok := submenus[profile]; ok {
				submenus[profile].Hide()
				submenusLock.Lock()
				delete(submenus, profile)
				submenusLock.Unlock()
				if _, ok := submenusToMenuItems[profile]; ok {
					submenusToMenuItemsLock.Lock()
					delete(submenusToMenuItems, profile)
					submenusToMenuItemsLock.Unlock()
				}
			}
		}

		if event.Op&fsnotify.Create == fsnotify.Create {
			profile := filepath.Base(event.Name)
			submenu := systray.AddSubMenu(strings.Title(profile))
			submenusLock.Lock()
			submenus[profile] = submenu
			submenusLock.Unlock()
			startMenu := submenu.AddSubMenuItem(START, "", 0)
			stopMenu := submenu.AddSubMenuItem(STOP, "", 0)
			consoleMenu := submenu.AddSubMenuItem(WEB_CONSOLE, "", 0)
			submenusToMenuItemsLock.Lock()
			ma := MenuAction{start: startMenu, stop: stopMenu, console: consoleMenu}
			submenusToMenuItems[profile] = ma
			submenusToMenuItemsLock.Unlock()

			go startStopHandler(icon.Running, profile, ma.start, START_PROFILE)

			go startStopHandler(icon.Stopped, profile, ma.stop, STOP_PROFILE)

			go webConsoleHandler(profile, ma.console)
		}

	}
}

// stopProfile stops a profile when clicked on the stop menuItem
func stopProfile(profileName string) error {
	minishiftBinary, _ := os.CurrentExecutable()
	if runtime.GOOS == "windows" {
		var stopCommandString = fmt.Sprintf(minishiftBinary + " stop --profile " + profileName)
		stopFilePath := filepath.Join(goos.TempDir(), "minishift_stop.bat")

		f, err := goos.Create(stopFilePath)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err = f.WriteString(stopCommandString); err != nil {
			return err
		}
		f.Close()

		posh := powershell.New()
		command := fmt.Sprintf("`Start-Process -FilePath %s", stopFilePath)
		_, _, err = posh.Execute(command)
		return err
	}

	if runtime.GOOS == "darwin" {
		var stopCommandString = fmt.Sprintf(minishiftBinary + " stop --profile " + profileName)
		stopFilePath := filepath.Join(goos.TempDir(), "minishift.stop")

		f, err := goos.Create(stopFilePath)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err = f.WriteString(stopCommandString); err != nil {
			return err
		}
		if err = f.Chmod(0744); err != nil {
			return err
		}
		f.Close()
		args := []string{"-F", "-a", "Terminal.app", stopFilePath}
		cmd, err := exec.LookPath("open")
		if err != nil {
			if glog.V(3) {
				fmt.Println("Could not find open in path")
				return fmt.Errorf("%v", err)
			}
		}
		command := exec.Command(cmd, args...)
		return command.Run()
	}
	return nil
}

// startProfile starts a profile when clicked on the start menuItem
func startProfile(profileName string) error {
	minishiftBinary, _ := os.CurrentExecutable()
	if runtime.GOOS == "windows" {
		var startCommandString = fmt.Sprintf(minishiftBinary + " start --profile " + profileName)
		startFilePath := filepath.Join(goos.TempDir(), "minishift_start.bat")

		f, err := goos.Create(startFilePath)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err = f.WriteString(startCommandString); err != nil {
			return err
		}
		f.Close()

		posh := powershell.New()
		command := fmt.Sprintf("Start-Process -FilePath %s", startFilePath)
		_, _, err = posh.Execute(command)
		return err
	}
	if runtime.GOOS == "darwin" {
		var startCommandString = fmt.Sprintf(minishiftBinary + " start --profile " + profileName)
		startFilePath := filepath.Join(goos.TempDir(), "minishift.start")

		f, err := goos.Create(startFilePath)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err = f.WriteString(startCommandString); err != nil {
			return err
		}
		if err = f.Chmod(0744); err != nil {
			return err
		}
		f.Close()

		args := []string{"-F", "-a", "Terminal.app", startFilePath}
		cmd, err := exec.LookPath("open")
		if err != nil {
			if glog.V(3) {
				fmt.Println("Could not find open in path")
				return fmt.Errorf("%v", err)
			}
		}
		command := exec.Command(cmd, args...)
		return command.Run()
	}
	return nil
}

// updateProfileStatus updates the menu bitmap to reflact the state of
// machine, green: running, red: stoppped, grey: does not exist
func updateProfileStatus() {
	for {
		time.Sleep(5 * time.Second)
		submenusLock.Lock()
		for k, v := range submenus {
			status := getStatus(k)
			if status == DOES_NOT_EXIST {
				v.AddBitmap(icon.DoesNotExist)
			}
			if status == RUNNING {
				v.AddBitmap(icon.Running)
			}
			if status == STOPPED {
				v.AddBitmap(icon.Stopped)
			}
		}
		submenusLock.Unlock()
	}
}

func startStopHandler(iconData []byte, submenu string, m *systray.MenuItem, action int) {
	var err error
	for {
		<-m.OnClickCh()
		if action == START_PROFILE {
			err = startProfile(submenu)
		} else {
			err = stopProfile(submenu)
		}
		if err == nil {
			submenusLock.Lock()
			submenus[submenu].AddBitmap(iconData)
			submenusLock.Unlock()
		}
	}
}

func webConsoleHandler(profile string, m *systray.MenuItem) {
	for {
		<-m.OnClickCh()
		api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
		host, _ := api.Load(profile)
		if cmdUtil.VMExists(api, profile) && cmdUtil.IsHostRunning(host.Driver) {
			ip, _ := host.Driver.GetIP()
			url := fmt.Sprintf("https://%s:%d/console", ip, constants.APIServerPort)
			browser.OpenURL(url)
		} else {
			continue
		}
		api.Close()
	}
}
