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

package minishiftdata

import (
	"encoding/json"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	minishiftDataPath string
)

type MinishiftData struct {
	OcPath string `json:"OcPath"`
}

// Create minishift-data.json with empty data or return with data if exists
func Create(machinesDirPath string) (*MinishiftData, error) {
	data := &MinishiftData{
		OcPath: "",
	}

	minishiftDataPath = filepath.Join(machinesDirPath, constants.MachineName, "minishift-data.json")
	if _, err := os.Stat(minishiftDataPath); os.IsNotExist(err) {
		if err := WriteMinishiftData(data); err != nil {
			return nil, err
		}

		return data, nil
	}

	return ReadMinishiftData()
}

func WriteMinishiftData(data *MinishiftData) error {
	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(minishiftDataPath, jsonData, 0644); err != nil {
		return err
	}

	return nil
}

func ReadMinishiftData() (*MinishiftData, error) {
	var data *MinishiftData

	raw, err := ioutil.ReadFile(minishiftDataPath)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(raw, &data)
	return data, nil
}
