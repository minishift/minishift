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

package cmd

import (
	"os"
	"reflect"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"io/ioutil"
	"path/filepath"

	"github.com/minishift/minishift/cmd/minishift/state"
	"github.com/minishift/minishift/cmd/testing/cli"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
)

var configTests = []cli.TestOption{
	{
		Name:          "v",
		ExpectedValue: "0",
	},
	{
		Name:          "v",
		ConfigValue:   `{ "v":"999" }`,
		ExpectedValue: "999",
	},
	{
		Name:          "v",
		FlagValue:     "0",
		ExpectedValue: "0",
	},
	{
		Name:          "v",
		EnvValue:      "123",
		ExpectedValue: "123",
	},
	{
		Name:          "v",
		FlagValue:     "3",
		ExpectedValue: "3",
	},
	// Flag should override config and env
	{
		Name:          "v",
		FlagValue:     "3",
		ConfigValue:   `{ "v": "222" }`,
		EnvValue:      "888",
		ExpectedValue: "3",
	},
	// Env should override config
	{
		Name:          "v",
		EnvValue:      "2",
		ConfigValue:   `{ "v": "999" }`,
		ExpectedValue: "2",
	},
	// Env should not override flags not on whitelist
	{
		Name:          "log_backtrace_at",
		EnvValue:      ":2",
		ExpectedValue: ":0",
	},
}

func TestPreRunDirectories(t *testing.T) {
	// Make sure we create the required directories.
	testDir, err := ioutil.TempDir("", "minishift-test-start-cmd-")
	// Need to create since Minishift config get created in root command
	os.Mkdir(filepath.Join(testDir, "config"), 0755)
	os.Mkdir(filepath.Join(testDir, "machines"), 0755)

	allInstanceConfigPath := filepath.Join(testDir, "config", "allinstances.json")
	minishiftConfig.AllInstancesConfig, err = minishiftConfig.NewAllInstancesConfig(allInstanceConfigPath)

	assert.NoError(t, err)
	defer os.RemoveAll(testDir)

	constants.Minipath = testDir
	cli.RunCommand(RootCmd.PersistentPreRun)

	dirPaths := reflect.ValueOf(*state.InstanceDirs)
	for i := 0; i < dirPaths.NumField(); i++ {
		dir := dirPaths.Field(i).Interface().(string)
		_, err := os.Stat(dir)
		assert.NoError(t, err, "Directory %s does not exist.", dir)
	}
}

func TestViperConfig(t *testing.T) {
	defer viper.Reset()
	err := cli.InitTestConfig(`{ "v": "999" }`)

	assert.NoError(t, err, "Failed to initialize the test config file")
	assert.Equal(t, viper.GetString("v"), "999", "Viper cannot read the test config file")
}

func TestViperAndFlags(t *testing.T) {
	restore := cli.HideEnv(t)
	defer restore(t)
	for _, testOption := range configTests {
		cli.SetOptionValue(t, testOption)
		setupViper()
		f := pflag.Lookup(testOption.Name)
		assert.NotNil(t, f)
		actual := f.Value.String()
		assert.Equal(t, testOption.ExpectedValue, actual)
		cli.UnsetValues(testOption)
	}
}

func TestInitializeProfile(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	Testdata := []struct {
		Args   []string
		Result string
	}{
		{[]string{"minishift", "profile", "set", "hello"}, "hello"},
		{[]string{"minishift", "profiles", "set", "hello"}, "hello"},
		{[]string{"minishift", "instance", "set", "hello"}, "hello"},
		{[]string{"minishift", "start", "--profile", "hello"}, "hello"},
		{[]string{"minishift", "ip", "--profile", "hello"}, "hello"},
		{[]string{"minishift", "status", "--profile", "hello"}, "hello"},
	}

	for _, testInput := range Testdata {
		os.Args = testInput.Args
		got := initializeProfile()
		assert.Equal(t, testInput.Result, got)
	}
}
