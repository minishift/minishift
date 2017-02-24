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
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/testing/cli"
	"io/ioutil"
	"path/filepath"
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
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(testDir)

	constants.Minipath = testDir
	cli.RunCommand(RootCmd.PersistentPreRun)

	for _, dir := range dirs {
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			t.Fatalf("Directory %s does not exist.", dir)
		}
	}
}

func TestViperConfig(t *testing.T) {
	defer viper.Reset()
	err := cli.InitTestConfig(`{ "v": "999" }`)
	if viper.GetString("v") != "999" || err != nil {
		t.Fatalf("Viper cannot read the test config file: %v", err)
	}
}

func TestViperAndFlags(t *testing.T) {
	restore := cli.HideEnv(t)
	defer restore(t)
	for _, testOption := range configTests {
		cli.SetOptionValue(t, testOption)
		setupViper()
		f := pflag.Lookup(testOption.Name)
		if f == nil {
			t.Fatalf("Cannot find the flag for %s", testOption.Name)
		}
		actual := f.Value.String()
		if actual != testOption.ExpectedValue {
			t.Errorf("pflag.Value(%s) => %s, wanted %s [%+v]", testOption.Name, actual, testOption.ExpectedValue, testOption)
		}
		cli.UnsetValues(testOption)
	}
}
