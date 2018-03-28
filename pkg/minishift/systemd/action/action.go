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

package action

type Action int

const (
	Start Action = iota
	Stop
	Restart
	Enable
	Disable
	Status
	//DaemonReload
)

func (action Action) String() string {
	actions := [...]string{
		"start",
		"stop",
		"restart",
		"enable",
		"disable",
		"status",
		//"daemon-reload",
	}

	if int(action) >= 0 && int(action) < len(actions) {
		return actions[action]
	}

	return ""
}
