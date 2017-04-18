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
	"errors"
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/mcnflag"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	name       = "Name"
	envVarName = "EnvVar"
)

func prepareDriverOptions(supportedFlags []mcnflag.Flag, explicitFlags map[string]interface{}) (drivers.DriverOptions, error) {
	driverOpts := rpcdriver.RPCFlags{
		Values: make(map[string]interface{}),
	}

	for _, flag := range supportedFlags {
		name := getField(flag, name)

		// first we try to apply from the explicit config
		value, found := explicitFlags[name]
		if found {
			driverOpts.Values[name] = value
			delete(explicitFlags, name)
		} else {
			driverOpts.Values[name] = flag.Default()
		}

		// now we potentially overwrite from the environment
		value, err := getEnvValue(flag)
		if err != nil {
			return nil, err
		}

		if value != "" {
			driverOpts.Values[name] = value
		}
	}

	if len(explicitFlags) != 0 {
		return nil, errors.New(fmt.Sprintf("Unused explicit driver options: %v", explicitFlags))
	}

	return &driverOpts, nil
}

func getField(flag mcnflag.Flag, name string) string {
	r := reflect.ValueOf(flag)
	return reflect.Indirect(r).FieldByName(name).String()
}

func getEnvValue(flag mcnflag.Flag) (interface{}, error) {
	envVarName := getField(flag, envVarName)
	if os.Getenv(envVarName) == "" {
		return "", nil
	}

	rawValue := os.Getenv(envVarName)
	var convertedValue interface{}
	var err error

	switch t := flag.(type) {
	case *mcnflag.StringFlag:
		convertedValue = rawValue
	case mcnflag.StringFlag:
		convertedValue = rawValue
	case mcnflag.StringSliceFlag:
		convertedValue = strings.Split(rawValue, " ")
	case *mcnflag.StringSliceFlag:
		convertedValue = strings.Split(rawValue, " ")
	case mcnflag.BoolFlag:
		convertedValue, err = strconv.ParseBool(rawValue)
	case *mcnflag.BoolFlag:
		convertedValue, err = strconv.ParseBool(rawValue)
	case mcnflag.IntFlag:
		convertedValue, err = strconv.Atoi(rawValue)
	case *mcnflag.IntFlag:
		convertedValue, err = strconv.Atoi(rawValue)
	default:
		return nil, errors.New(fmt.Sprintf("Unexpected flag type %v", t))
	}

	if err != nil {
		return nil, err
	}

	return convertedValue, nil
}
