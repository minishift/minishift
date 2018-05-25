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
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

type KvpRecord struct {
	Key   [MAX_KEY_SIZE]byte
	Value [MAX_VALUE_SIZE]byte
}

func (record *KvpRecord) GetKey() string {
	return strings.Trim(string(record.Key[:MAX_KEY_SIZE]), "\x00")
}

func (record *KvpRecord) GetValue() string {
	return strings.Trim(string(record.Value[:MAX_VALUE_SIZE]), "\x00")
}

const (
	MAX_KEY_SIZE     = 512
	MAX_VALUE_SIZE   = 2048
	DEFAULT_POOLNAME = "/var/lib/hyperv/.kvp_pool_0"
)

func readNextBytes(file *os.File, number int) ([]byte, error) {
	bytes := make([]byte, number)

	_, err := file.Read(bytes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func getKvpRecords(poolFile string) []KvpRecord {
	file, err := os.Open(poolFile)
	if err != nil {
		fmt.Println("Error opening pool")
		os.Exit(3)
	}

	var records []KvpRecord

	for {
		record := KvpRecord{}
		data, err := readNextBytes(file, MAX_KEY_SIZE+MAX_VALUE_SIZE)
		buffer := bytes.NewBuffer(data)
		err = binary.Read(buffer, binary.LittleEndian, &record)
		if err == io.EOF {
			break
		}

		records = append(records, record)
	}

	return records
}

func GetKvpRecordByKey(key string) *KvpRecord {
	for _, record := range getKvpRecords(DEFAULT_POOLNAME) {
		if strings.EqualFold(record.GetKey(), key) {
			return &record
		}
	}
	return nil
}

func GetAllKvpRecords() []KvpRecord {
	return getKvpRecords(DEFAULT_POOLNAME)
}
