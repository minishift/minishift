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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const invalidSleepTimeError = "Unable to extract sleep time from cmd: %s"

var sleepRegexp = regexp.MustCompile(`sleep (\d+)`)

type SleepCommand struct {
	*defaultCommand
}

func NewSleepCommand(command string) *SleepCommand {
	defaultCommand := &defaultCommand{rawCommand: command}
	sleepCommand := &SleepCommand{defaultCommand}
	defaultCommand.fn = sleepCommand.doExecute
	return sleepCommand
}

func (c *SleepCommand) doExecute(ec *ExecutionContext) error {
	duration, err := c.getSleepTime()
	if err != nil {
		return err
	}

	fmt.Print(".")
	time.Sleep(duration)
	return nil
}

func (c *SleepCommand) getSleepTime() (time.Duration, error) {
	sleepSlice := sleepRegexp.FindStringSubmatch(c.rawCommand)

	if sleepSlice == nil {
		return -1, errors.New(fmt.Sprintf(invalidSleepTimeError, c.rawCommand))
	}
	duration, _ := strconv.Atoi(sleepSlice[1])

	return time.Duration(duration) * time.Second, nil
}
