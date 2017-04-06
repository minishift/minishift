/*
Copyright (C) 2016 Red Hat, Inc.

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

package atexit

import (
	"fmt"
	"os"

	"github.com/golang/glog"
)

var exitHandlers = []func(code int) int{}

// Exit runs all registered exit handlers and then exits the program with the specified exit code using os.Exit
func Exit(code int) {
	code = runHandlers(code)
	os.Exit(code)
}

// ExitWithMessage exit and print the error message.
func ExitWithMessage(code int, msg string) {
	if code == 0 {
		fmt.Fprintln(os.Stdout, msg)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
	Exit(code)
}

// Register registers an exit handler function which is run when Exit is called
func RegisterExitHandler(exitHandler func(code int) int) {
	exitHandlers = append(exitHandlers, exitHandler)
}

func ClearExitHandler() {
	exitHandlers = []func(code int) int{}
}

func runHandlers(code int) int {
	var exitCode int
	for _, handler := range exitHandlers {
		exitCode = runHandler(handler, code)
	}
	return exitCode
}

func runHandler(exitHandler func(code int) int, code int) int {
	defer func() {
		err := recover()
		if err != nil {
			glog.Exitf("Error running exit handler:", err)
		}
	}()

	return exitHandler(code)
}
