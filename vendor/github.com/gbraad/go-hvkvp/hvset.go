/*
Copyright (C) 2017 Gerard Braad <me@gbraad.nl>

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

package hvkvp

import (
	"fmt"
)

func NewMachineKVPRecord(machineName string, key string, value string) *MachineKVPRecord {
	return &MachineKVPRecord{
		Machine: machineName,
		Key:     key,
		Value:   value,
		Pool:    0,
	}
}

func NewMachineKVPCommand(record *MachineKVPRecord) string {
	return prepareKeyValuePairCommand(record)
}

type MachineKVPRecord struct {
	Machine string
	Key     string
	Value   string
	Pool    int
}

func prepareKeyValuePairCommand(record *MachineKVPRecord) string {
	if record == nil {
		return ""
	}

	return (`
$vmMgmt = Get-WmiObject -Namespace root\virtualization\v2 -Class Msvm_VirtualSystemManagementService
$vm = Get-WmiObject -Namespace root\virtualization\v2 -Class Msvm_ComputerSystem -Filter {` + fmt.Sprintf("ElementName = '%s'", record.Machine) + `}
$kvpDataItem = ([WMIClass][String]::Format("\\{0}\{1}:{2}", $vmMgmt.ClassPath.Server, $vmMgmt.ClassPath.NamespacePath, "Msvm_KvpExchangeDataItem")).CreateInstance()
` + fmt.Sprintf("$kvpDataItem.Name = '%s'", record.Key) + `
` + fmt.Sprintf("$kvpDataItem.Data = '%s'", record.Value) + `
` + fmt.Sprintf("$kvpDataItem.Source = '%d'", record.Pool) + `
$vmMgmt.RemoveKvpItems($vm, $kvpDataItem.PSBase.GetText(1))
$result = $vmMgmt.AddKvpItems($vm, $kvpDataItem.PSBase.GetText(1))
$result.ReturnValue
	`)
}
