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

package cluster

import (
	"github.com/docker/machine/libmachine/drivers"
	"os"
	"reflect"
)

func setDriverOptionsFromEnvironment(d drivers.Driver) error {
	supportedFlags := d.GetCreateFlags()

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{},
		CreateFlags: supportedFlags,
	}

	for _, flag := range supportedFlags {
		r := reflect.ValueOf(flag)
		flagName := reflect.Indirect(r).FieldByName("Name").String()
		flagEnvName := reflect.Indirect(r).FieldByName("EnvVar").String()

		if os.Getenv(flagEnvName) != "" {
			checkFlags.FlagsValues[flagName] = os.Getenv(flagEnvName)
		}
	}

	if err := d.SetConfigFromFlags(checkFlags); err != nil {
		return err
	}

	return nil
}
