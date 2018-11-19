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
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/progressdots"
	"regexp"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/provision"
	minishiftStrings "github.com/minishift/minishift/pkg/util/strings"
)

type RegistrationParameters struct {
	Username               string
	Password               string
	IsTtySupported         bool
	GetUsernameInteractive func(message string) string
	GetPasswordInteractive func(message string) string
	GetPasswordKeyring     func(username string) (string, error)
	SetPasswordKeyring     func(username, password string) error
}

// Detect supported Registrator
func NeedRegistration(host *host.Host) (bool, error) {
	commander := provision.GenericSSHCommander{Driver: host.Driver}
	_, supportRegistration, err := DetectRegistrator(commander)
	return supportRegistration, err
}

type RegistrationHostActionFunc func(param *RegistrationParameters) error

func doRegistrationHostAction(actionMessage string, registrationAction RegistrationHostActionFunc, param *RegistrationParameters) (bool, error) {
	fmt.Println(actionMessage)
	if err := registrationAction(param); err != nil {
		// error occured during action
		return false, err
	}

	// was successful
	return true, nil
}

// Register host VM
func RegisterHostVM(host *host.Host, param *RegistrationParameters) (bool, error) {
	commander := provision.GenericSSHCommander{Driver: host.Driver}
	registrator, supportRegistration, err := DetectRegistrator(commander)
	if !supportRegistration {
		log.Debug("Distribution doesn't support registration")
	}
	if err != nil && err != ErrDetectionFailed {
		return supportRegistration, err
	}
	if err != nil && err != ErrDetectionFailed {
		// failed
		return supportRegistration, err
	}
	if err == ErrDetectionFailed || registrator == nil {
		// does not support registration
		return supportRegistration, nil
	}

	return doRegistrationHostAction(
		"-- Registering machine using subscription-manager",
		registrator.Register,
		param)
}

// Unregister host VM
func UnregisterHostVM(host *host.Host, param *RegistrationParameters) (bool, error) {
	commander := provision.GenericSSHCommander{Driver: host.Driver}
	registrator, supportRegistration, err := DetectRegistrator(commander)
	if !supportRegistration {
		log.Debug("Distribution doesn't support unregistration")
	}
	if err != nil && err != ErrDetectionFailed {
		// failed
		return supportRegistration, err
	}
	if err == ErrDetectionFailed || registrator == nil {
		// does not support unregistration
		return supportRegistration, nil
	}

	return doRegistrationHostAction(
		"Unregistering machine",
		registrator.Unregister,
		param)
}

func RSHMRegisterAndLoginToRedhatRegistry(sshCommander provision.SSHCommander, param *RegistrationParameters, isRHEL bool) error {
	var err error
	keyringDocsLink := "https://docs.okd.io/latest/minishift/troubleshooting/troubleshooting-misc.html#Remove-password-from-keychain"
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
				fmt.Printf("   Retriving password from keychain ...")
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
		_, err = sshCommander.SSHCommand(dockerLoginCommand)
		progressDots.Stop()
		if err == nil {
			fmt.Println(" OK ")
		} else {
			fmt.Println(" FAIL ")
		}

		if isRHEL {
			fmt.Printf("   Registration in progress ")
			// Initialize progressDots channel again for registration
			progressDots = progressdots.New()
			progressDots.Start()
			startTime := time.Now()
			// start timed SSH command to register
			_, err = sshCommander.SSHCommand(subscriptionCommand)
			progressDots.Stop()
			if err == nil {
				fmt.Print(" OK ")
			} else {
				fmt.Print(" FAIL ")
			}
			fmt.Printf("[%s]\n", util.TimeElapsed(startTime, true))
		}

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
	return nil
}

func redactPassword(msg string) string {
	pattern := regexp.MustCompile(`--password .*`)
	msgWithRedactedPass := pattern.ReplaceAllString(msg, "--password ********")
	return msgWithRedactedPass
}
