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
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/provision"
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/progressdots"
	minishiftStrings "github.com/minishift/minishift/pkg/util/strings"
)

const (
	RHSMName       = "   Red Hat Developers or Red Hat Subscription Management (RHSM)"
	RedHatRegistry = "registry.redhat.io"
)

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

// CompatibleWithDistribution returns true if system supports registration with RHSM
func (registrator *RedHatRegistrator) CompatibleWithDistribution(osReleaseInfo *provision.OsRelease) bool {
	if osReleaseInfo.ID != "rhel" {
		return false
	}
	if _, err := registrator.SSHCommand("sudo -E subscription-manager"); err != nil {
		return false
	}
	return true
}

// Register attempts to register the system with RHSM and registry.redhat.io
func (registrator *RedHatRegistrator) Register(param *RegistrationParameters) error {
	keyringDocsLink := "https://docs.okd.io/latest/minishift/troubleshooting/troubleshooting-misc.html#Remove-password-from-keychain"
	if isRegistered, err := registrator.IsRegistered(); !isRegistered && err == nil {
		for i := 1; i < 4; i++ {
			// request username (disallow empty value)
			if param.Username == "" {
				// Check if Terminal tty supported or not
				if !param.IsTtySupported {
					return fmt.Errorf(("Not a tty supported terminal, Retries are disabled"))
				}
				for param.Username == "" {
					param.Username = param.GetUsernameInteractive(RHSMName + " username")
				}
			}
			// request password (disallow empty value)
			if param.Password == "" {
				if i == 1 {
					fmt.Printf("   Retrieving password from keychain ...")
					param.Password, err = param.GetPasswordKeyring(param.Username)
					if err != nil {
						fmt.Println(" FAIL ")
					} else {
						fmt.Println(" OK ")
					}
				}

				for param.Password == "" {
					param.Password = param.GetPasswordInteractive(RHSMName + " password")
					fmt.Printf("   Storing password in keychain ...")
					err := param.SetPasswordKeyring(param.Username, param.Password)
					if err != nil {
						fmt.Println(" FAIL ")
					} else {
						fmt.Println(" OK ")
					}
					if glog.V(3) {
						fmt.Println(err)
					}
				}
			}

			// prepare subscription command
			subscriptionCommand := fmt.Sprintf("sudo -E subscription-manager register --auto-attach "+
				"--username %s "+
				"--password '%s' ",
				param.Username,
				minishiftStrings.EscapeSingleQuote(param.Password))

			// prepare docker login command for registry.redhat.io
			dockerLoginCommand := fmt.Sprintf("docker login --username %s --password %s %s",
				param.Username,
				minishiftStrings.EscapeSingleQuote(param.Password),
				RedHatRegistry)

			fmt.Printf("   Login to %s in progress ", RedHatRegistry)
			// Initialize progressDots channel
			progressDots := progressdots.New()
			progressDots.Start()
			// Login to docker registry
			_, err = registrator.SSHCommand(dockerLoginCommand)
			progressDots.Stop()
			if err == nil {
				fmt.Println(" OK ")
			} else {
				fmt.Println(" FAIL ")
			}

			fmt.Printf("   Registration in progress ")
			// Initialize progressDots channel again for registration
			progressDots = progressdots.New()
			progressDots.Start()
			startTime := time.Now()
			// start timed SSH command to register
			_, err = registrator.SSHCommand(subscriptionCommand)
			progressDots.Stop()
			if err == nil {
				fmt.Print(" OK ")
			} else {
				fmt.Print(" FAIL ")
			}
			fmt.Printf("[%s]\n", util.TimeElapsed(startTime, true))

			if err == nil {
				return nil
			}

			// general error when registration fails
			if strings.Contains(err.Error(), "Invalid username or password") {
				fmt.Printf("   Invalid username or password. To delete stored password refer to %s. Retry: %d\n", keyringDocsLink, i)
			} else {
				return errors.New(redactPassword(err.Error()))
			}

			// always reset credentials
			param.Username = ""
			param.Password = ""
		}
		if err != nil {
			return errors.New(redactPassword(err.Error()))
		}
	}
	return nil
}

// Unregister attempts to unregister the system from RHSM
func (registrator *RedHatRegistrator) Unregister(param *RegistrationParameters) error {
	if isRegistered, err := registrator.IsRegistered(); isRegistered {
		if _, err := registrator.SSHCommand(
			"sudo -E subscription-manager unregister"); err != nil {
			return err
		}
		if _, err := registrator.SSHCommand(
			"sudo -E subscription-manager clean"); err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

// isRegistered returns registration state of RHSM or errors when undetermined
func (registrator *RedHatRegistrator) IsRegistered() (bool, error) {
	if output, err := registrator.SSHCommand("sudo -E subscription-manager list"); err != nil {
		return false, err
	} else {
		if !strings.Contains(output, "Unknown") {
			return true, nil
		}
		return false, nil
	}
}

func redactPassword(msg string) string {
	pattern := regexp.MustCompile(`--password .*`)
	msgWithRedactedPass := pattern.ReplaceAllString(msg, "--password ********")
	return msgWithRedactedPass
}
