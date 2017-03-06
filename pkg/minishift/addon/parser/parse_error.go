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

package parser

type ParseError interface {
	error
	AddonName() string
	AddonDir() string
}

type DefaultParseError struct {
	msg       string
	addonDir  string
	addonName string
}

func NewParseError(msg string, name string, dir string) *DefaultParseError {
	return &DefaultParseError{
		msg:       msg,
		addonDir:  dir,
		addonName: name,
	}
}

func (e *DefaultParseError) Error() string { return e.msg }

func (e *DefaultParseError) AddonName() string {
	if e.addonName != "" {
		return e.addonName
	} else {
		return ""
	}
}

func (e *DefaultParseError) AddonDir() string {
	if e.addonDir != "" {
		return e.addonDir
	} else {
		return ""
	}
}
