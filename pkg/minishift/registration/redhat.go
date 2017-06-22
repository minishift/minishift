/*
Copyright (C) 2016 Red Hat, Inc.

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

package registration

import (
	"fmt"
	"strings"

	"github.com/briandowns/spinner"
	"github.com/docker/machine/libmachine/provision"
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/util"
	minishiftStrings "github.com/minishift/minishift/pkg/util/strings"
	"os"
	"time"
)

const SPINNER_SPEED = 200 // In milliseconds

func init() {
	Register("Redhat", &RegisteredRegistrator{
		New: NewRedHatRegistrator,
	})
}

func NewRedHatRegistrator(c provision.SSHCommander) Registrator {
	return &RedHatRegistrator{
		SSHCommander: c,
	}
}

type RedHatRegistrator struct {
	provision.SSHCommander
}

func (registrator *RedHatRegistrator) CompatibleWithDistribution(osReleaseInfo *provision.OsRelease) bool {
	if osReleaseInfo.ID != "rhel" {
		return false
	}
	if _, err := registrator.SSHCommand("sudo -E subscription-manager"); err != nil {
		return false
	} else {
		return true
	}
}

func (registrator *RedHatRegistrator) SSHCommand(cmd string) (string, error) {
	startTime := time.Now()
	out, err := registrator.SSHCommander.SSHCommand(cmd)
	elapsed := time.Since(startTime)

	if glog.V(2) {
		fmt.Printf("Command \"%s\" took %v\n", cmd, util.FriendlyDuration(elapsed).String())
	}

	return out, err
}

func (registrator *RedHatRegistrator) Register(param *RegistrationParameters) error {
	output, err := registrator.SSHCommand("sudo -E subscription-manager version")
	if err != nil {
		return err
	}
	if strings.Contains(output, "not registered") {
		spinnerView := spinner.New(spinner.CharSets[9], SPINNER_SPEED*time.Millisecond)
		for i := 1; i < 4; i++ {
			if param.Username == "" {
				param.Username = param.GetUsernameInteractive("Red Hat Developers or Red Hat Subscription Management (RHSM) username")
			}
			if param.Password == "" {
				param.Password = param.GetPasswordInteractive("Red Hat Developers or Red Hat Subscription Management (RHSM) password")
			}
			subscriptionCommand := fmt.Sprintf("sudo -E subscription-manager register --auto-attach "+
				"--username %s "+
				"--password '%s' ", param.Username, minishiftStrings.EscapeSingleQuote(param.Password))
			spinnerView.Start()
			startTime := time.Now()
			_, err = registrator.SSHCommand(subscriptionCommand)
			spinnerView.Stop()
			if err == nil {
				fmt.Print("Registration successful ")
			} else {
				fmt.Print("Registration unsuccessful ")
			}
			util.TimeTrack(startTime, os.Stdout, true)
			if err == nil {
				return nil
			}
			if strings.Contains(err.Error(), "Invalid username or password") {
				fmt.Println("Invalid username or password Retry: ", i)
				param.Username = ""
				param.Password = ""
			} else {
				return err
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (registrator *RedHatRegistrator) Unregister(param *RegistrationParameters) error {
	if output, err := registrator.SSHCommand("sudo -E subscription-manager version"); err != nil {
		return err
	} else {
		if !strings.Contains(output, "not registered") {
			if _, err := registrator.SSHCommand(
				"sudo -E subscription-manager unregister"); err != nil {
				return err
			}
		}
	}
	return nil
}
