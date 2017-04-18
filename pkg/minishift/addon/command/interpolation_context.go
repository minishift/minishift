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

package command

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
)

// InterpolationContext allows to interpolate variables within commands
type InterpolationContext interface {
	// AddToContext adds the specified value under the the specified key for command interpolation
	AddToContext(key string, value string) error

	// RemoveFromContext removes the specified value from the context
	RemoveFromContext(key string) error

	// Interpolate the cmd in the current context
	Interpolate(cmd string) string
}

type defaultInterpolationContext struct {
	InterpolationContext

	context map[string]regExpAndValue
}

type regExpAndValue struct {
	regexp      *regexp.Regexp
	subsitution string
}

// NewInterpolationContext creates a new interpolation context
func NewInterpolationContext() InterpolationContext {
	context := make(map[string]regExpAndValue)
	return &defaultInterpolationContext{context: context}
}

func (c *defaultInterpolationContext) AddToContext(key string, value string) error {
	r, err := regexp.Compile(fmt.Sprintf("#{%s}", key))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to add %s/%s to interpolation context", key, value))
	}
	c.context[key] = regExpAndValue{regexp: r, subsitution: value}
	return nil
}

func (c *defaultInterpolationContext) RemoveFromContext(key string) error {
	delete(c.context, key)
	return nil
}

func (c *defaultInterpolationContext) Interpolate(cmd string) string {
	for _, v := range c.context {
		cmd = v.regexp.ReplaceAllString(cmd, v.subsitution)
	}
	return cmd
}
