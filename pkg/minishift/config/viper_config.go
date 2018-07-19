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

package config

import (
	"encoding/json"
	"fmt"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"io"
	"os"
)

type ViperConfig map[string]interface{}

// ReadConfig reads the config from $MINISHIFT_HOME/config/config.json file
func ReadViperConfig(configfile string) (ViperConfig, error) {
	f, err := os.Open(configfile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("Cannot open file '%s': %s", constants.ConfigFile, err)
	}
	var m ViperConfig
	m, err = Decode(f)
	if err != nil {
		return nil, fmt.Errorf("Cannot decode config '%s': %s", constants.ConfigFile, err)
	}

	return m, nil
}

// Writes a config to the $MINISHIFT_HOME/config/config.json file
func WriteViperConfig(configfile string, m ViperConfig) error {
	f, err := os.Create(configfile)
	if err != nil {
		return fmt.Errorf("Cannot create file '%s': %s", constants.ConfigFile, err)
	}
	defer f.Close()
	err = Encode(f, m)
	if err != nil {
		return fmt.Errorf("Cannot encode config '%s': %s", constants.ConfigFile, err)
	}
	return nil
}

func Decode(r io.Reader) (ViperConfig, error) {
	var data ViperConfig
	err := json.NewDecoder(r).Decode(&data)
	return data, err
}

func Encode(w io.Writer, m ViperConfig) error {
	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}

	_, err = w.Write(b)

	return err
}
