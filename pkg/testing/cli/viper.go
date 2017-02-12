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

package cli

import (
	"bytes"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
	"testing"
)

type TestOption struct {
	Name          string
	EnvValue      string
	ConfigValue   string
	FlagValue     string
	ExpectedValue string
}

func RunCommand(f func(*cobra.Command, []string)) {
	cmd := cobra.Command{}
	var args []string
	f(&cmd, args)
}

func SetOptionValue(t *testing.T, testOption TestOption) {
	if testOption.FlagValue != "" {
		pflag.Set(testOption.Name, testOption.FlagValue)
	}
	if testOption.EnvValue != "" {
		s := strings.Replace(getEnvVarName(testOption.Name), "-", "_", -1)
		os.Setenv(s, testOption.EnvValue)
	}
	if testOption.ConfigValue != "" {
		err := InitTestConfig(testOption.ConfigValue)
		if err != nil {
			t.Fatalf("Config %s not read correctly: %v", testOption.ConfigValue, err)
		}
	}
}

func UnsetValues(testOption TestOption) {
	var f = pflag.Lookup(testOption.Name)
	if f != nil {
		f.Value.Set(f.DefValue)
		f.Changed = false
	}

	os.Unsetenv(getEnvVarName(testOption.Name))

	viper.Reset()
}

// Temporarily unsets the env variables for the test cases
// returns a function to reset them to their initial values
func HideEnv(t *testing.T) func(t *testing.T) {
	envs := make(map[string]string)
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, constants.MiniShiftEnvPrefix) {
			line := strings.Split(env, "=")
			key, val := line[0], line[1]
			envs[key] = val
			t.Logf("TestConfig: Unsetting %s=%s for unit test!", key, val)
			os.Unsetenv(key)
		}
	}
	return func(t *testing.T) {
		for key, val := range envs {
			t.Logf("TestConfig: Finished test, Resetting Env %s=%s", key, val)
			os.Setenv(key, val)
		}
	}
}

func InitTestConfig(config string) error {
	viper.SetConfigType("json")
	r := bytes.NewReader([]byte(config))
	return viper.ReadConfig(r)
}

func getEnvVarName(name string) string {
	return constants.MiniShiftEnvPrefix + "_" + strings.ToUpper(name)
}
