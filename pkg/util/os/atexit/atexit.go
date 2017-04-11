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

const ExitHandlerPanicMessage = "At least on exit handler vetoed to exit program execution"

// exitHandlers keeps track of the list of registered exit handlers. Handlers are applied in the order defined in this list.
var exitHandlers = []func(code int) bool{}

// Exit runs all registered exit handlers and then exits the program with the specified exit code using os.Exit.
func Exit(code int) {
	veto := runHandlers(code)
	if veto {
		panic(ExitHandlerPanicMessage)
	}
	os.Exit(code)
}

// ExitWithMessage runs all registered exit handlers, prints the specified message and then exits the program with the specified exit code.
// If the exit code is 0, the message is prints to stdout, otherwise to stderr.
func ExitWithMessage(code int, msg string) {
	if code == 0 {
		fmt.Fprintln(os.Stdout, msg)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
	Exit(code)
}

// Register registers an exit handler function which is run when Exit is called
func RegisterExitHandler(exitHandler func(code int) bool) {
	exitHandlers = append(exitHandlers, exitHandler)
}

// ClearExitHandler clears all registered exit handlers
func ClearExitHandler() {
	exitHandlers = []func(code int) bool{}
}

// runHandlers runs all registered exit handlers, passing on the intended exit code.
// Handlers can veto to exit the program. If at least one exit handlers casts a veto, the program does panic instead of exiting
// allowing for a potential recovering.
func runHandlers(code int) bool {
	var veto bool
	for _, handler := range exitHandlers {
		veto = veto || runHandler(handler, code)
	}
	return veto
}

// runHandler runs the single specified exit handler, returning whether this handler vetos the exit or not.
func runHandler(exitHandler func(code int) bool, code int) bool {
	defer func() {
		err := recover()
		if err != nil {
			glog.Exitf("Error running exit handler:", err)
		}
	}()

	return exitHandler(code)
}
