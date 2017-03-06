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

package addon

import (
	"errors"
	"fmt"
)

var requiredMetaTags = []string{"Name", "Description"}

// AddOnMeta defines a set of meta data for a AddOn. Name and Description are required. Others are optional.
type AddOnMeta interface {
	Name() string
	Description() string
	GetValue(key string) string
}

type DefaultAddOnMeta struct {
	headers map[string]string
}

func NewAddOnMeta(headers map[string]string) (AddOnMeta, error) {
	for _, tag := range requiredMetaTags {
		if headers[tag] == "" {
			return nil, errors.New(fmt.Sprintf("Metadata does not contain an mandatory entry for '%s'", tag))
		}
	}

	metaData := &DefaultAddOnMeta{headers: headers}
	return metaData, nil
}

func (meta *DefaultAddOnMeta) String() string {
	return fmt.Sprintf("%#v", meta)
}

func (meta *DefaultAddOnMeta) Name() string {
	return meta.headers[requiredMetaTags[0]]
}

func (meta *DefaultAddOnMeta) Description() string {
	return meta.headers[requiredMetaTags[1]]
}

func (meta *DefaultAddOnMeta) GetValue(key string) string {
	return meta.headers[key]
}
