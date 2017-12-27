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

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine"

	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/pkg/minikube/cluster"
)

// Runs all the validation or callback functions and collects errors
func run(name string, value string, fns []setFn) error {
	var errors []error
	for _, fn := range fns {
		err := fn(name, value)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("%v", errors)
	} else {
		return nil
	}
}

func findSetting(name string) (Setting, error) {
	for _, s := range settingsList {
		if name == s.Name {
			return s, nil
		}
	}
	return Setting{}, fmt.Errorf("Cannot find property name '%s'", name)
}

// Set Functions

func SetString(m MinishiftConfig, name string, val string) error {
	m[name] = val
	return nil
}

func SetInt(m MinishiftConfig, name string, val string) error {
	i, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	m[name] = i
	return nil
}

func SetBool(m MinishiftConfig, name string, val string) error {
	b, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	m[name] = b
	return nil
}

func SetSlice(m MinishiftConfig, name string, val string) error {
	var tmpSlice []string
	if val != "" {
		for _, v := range strings.Split(val, ",") {
			tmpSlice = append(tmpSlice, strings.TrimSpace(v))
		}
	}
	m[name] = tmpSlice
	return nil
}

func RequiresRestartMsg(name string, value string) error {
	api := libmachine.NewClient(state.InstanceDirs.Home, state.InstanceDirs.Certs)
	defer api.Close()

	_, err := cluster.CheckIfApiExistsAndLoad(api)
	if err != nil {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("No Minishift instance exists. New '%s' setting will be applied on next 'minishift start'", name))
	} else {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("You currently have an existing Minishift instance. "+
			"Changes to the '%s' setting are only applied when a new Minishift instance is created.\n"+
			"To let the configuration changes take effect, "+
			"you must delete the current instance with 'minishift delete' "+
			"and then start a new one with 'minishift start'.", name))
	}
	return nil
}
