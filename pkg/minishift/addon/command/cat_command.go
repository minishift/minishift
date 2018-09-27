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

package command

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type CatCommand struct {
	*defaultCommand
}

func NewCatCommand(command string, ignoreError bool, outputVariable string) *CatCommand {
	defaultCommand := &defaultCommand{rawCommand: command, ignoreError: ignoreError, outputVariable: outputVariable}
	catCommand := &CatCommand{defaultCommand}
	defaultCommand.fn = catCommand.doExecute
	return catCommand
}

func (c *CatCommand) doExecute(ec *ExecutionContext, ignoreError bool, outputVariable string) error {
	var (
		err   error
		f     *os.File
		fInfo os.FileInfo
	)

	// split off the actual 'cat' command. As we need to get file name of remaining string.
	catString := strings.TrimPrefix(c.rawCommand, "cat")
	fileName := strings.Replace(catString, " ", "", 1)

	if fInfo, err = os.Stat(fileName); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("File %s doesn't exist", fileName))
	}

	f, err = os.Open(fileName)
	defer f.Close()
	fileSize := fInfo.Size()
	buffer := make([]byte, fileSize)
	_, err = f.Read(buffer)

	if err != nil {
		return errors.New(fmt.Sprintf("Error executing command '%s':", err.Error()))
	}

	if outputVariable != "" {
		ec.AddToContext(outputVariable, strings.TrimSpace(string(buffer)))
		return nil
	}
	fmt.Printf("%s", string(buffer))
	return nil
}
