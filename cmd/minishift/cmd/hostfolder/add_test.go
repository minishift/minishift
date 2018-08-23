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

package hostfolder

import (
	"fmt"
	"github.com/minishift/minishift/cmd/testing/cli"
	"github.com/minishift/minishift/pkg/minishift/config"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func Test_host_folder_name_required(t *testing.T) {
	var err error
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	config.InstanceConfig, err = config.NewInstanceConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	config.AllInstancesConfig, err = config.NewAllInstancesConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	assert.NoError(t, err, "Unexpected error setting instance config")

	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)
	defer viper.Reset()

	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 1, usage))
	addHostFolder(nil, nil)
}

func Test_host_folder_name_not_empty(t *testing.T) {
	var err error
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	config.InstanceConfig, err = config.NewInstanceConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	config.AllInstancesConfig, err = config.NewAllInstancesConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	assert.NoError(t, err, "Unexpected error setting instance config")

	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)
	defer viper.Reset()

	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 1, usage))
	addHostFolder(nil, []string{" "})
}

func Test_source_required(t *testing.T) {
	var err error
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	config.InstanceConfig, err = config.NewInstanceConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	config.AllInstancesConfig, err = config.NewAllInstancesConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	assert.NoError(t, err, "Unexpected error setting instance config")

	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)
	defer viper.Reset()

	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 1, noSource))
	addHostFolder(nil, []string{"foo"})
}

func Test_target_required(t *testing.T) {
	var err error
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	config.InstanceConfig, err = config.NewInstanceConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	config.AllInstancesConfig, err = config.NewAllInstancesConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	assert.NoError(t, err, "Unexpected error setting instance config")

	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)
	defer viper.Reset()

	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 1, noTarget))
	source = "/home/johndoe"
	addHostFolder(nil, []string{"foo"})
}

func Test_username_required(t *testing.T) {
	var err error
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	config.InstanceConfig, err = config.NewInstanceConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	config.AllInstancesConfig, err = config.NewAllInstancesConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	assert.NoError(t, err, "Unexpected error setting instance config")

	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)
	defer viper.Reset()

	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 1, noUserName))
	source = "/home/johndoe"
	target = "/var/tmp"
	addHostFolder(nil, []string{"foo"})
}

func Test_password_required(t *testing.T) {
	var err error
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	config.InstanceConfig, err = config.NewInstanceConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	config.AllInstancesConfig, err = config.NewAllInstancesConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	assert.NoError(t, err, "Unexpected error setting instance config")

	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)
	defer viper.Reset()

	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 1, noPassword))
	source = "/home/johndoe"
	target = "/var/tmp"
	options = "username=johndoe"
	addHostFolder(nil, []string{"foo"})
}

func Test_domain_required(t *testing.T) {
	var err error
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	config.InstanceConfig, err = config.NewInstanceConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	config.AllInstancesConfig, err = config.NewAllInstancesConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	assert.NoError(t, err, "Unexpected error setting instance config")

	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)
	defer viper.Reset()

	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 1, noDomain))
	source = "/home/johndoe"
	target = "/var/tmp"
	options = "username=johndoe,password==,123"
	addHostFolder(nil, []string{"foo"})
}

func Test_unknown_host_folder_type_exits(t *testing.T) {
	var err error
	tmpMinishiftHomeDir := cli.SetupTmpMinishiftHome(t)
	config.InstanceConfig, err = config.NewInstanceConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	config.AllInstancesConfig, err = config.NewAllInstancesConfig(filepath.Join(tmpMinishiftHomeDir, "config"))
	assert.NoError(t, err, "Unexpected error setting instance config")

	tee := cli.CreateTee(t, true)
	defer cli.TearDown(tmpMinishiftHomeDir, tee)
	defer viper.Reset()

	atexit.RegisterExitHandler(cli.VerifyExitCodeAndMessage(t, tee, 1, fmt.Sprintf(unknownType, "snafu")))
	source = "/home/johndoe"
	target = "/var/tmp"
	shareType = "snafu"
	addHostFolder(nil, []string{"foo"})
}
