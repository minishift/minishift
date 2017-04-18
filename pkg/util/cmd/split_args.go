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

package cmd

import (
	"strings"
	"unicode"
)

// This is a '\'
const escapeRune = rune(92)

// SplitCmdString takes a single line command string and splits it into individual arguments to be used with
// exec.Command. It does so by honouring quotes within the original command
func SplitCmdString(cmd string) []string {
	lastQuote := rune(0)
	consecutiveEscapeCount := 0
	splitWithQuotes := func(c rune) bool {
		switch {
		// don't split if the character matches the "last matched quote"
		case c == lastQuote:
			if consecutiveEscapeCount%2 == 0 {
				lastQuote = rune(0)
			}
			consecutiveEscapeCount = 0
			return false
		// found an escape character, don't split
		case c == escapeRune:
			consecutiveEscapeCount += 1
			return false
		// don't split if we are still in a quoted string
		case lastQuote != rune(0):
			consecutiveEscapeCount = 0
			return false
		// don't split if the character is a quote type rune, and set the lastQuote value to that
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			consecutiveEscapeCount = 0
			return false
		// split the string if it is a space and none of the other conditionals are satisfied
		default:
			consecutiveEscapeCount = 0
			return unicode.IsSpace(c)
		}
	}
	result := strings.FieldsFunc(cmd, splitWithQuotes)
	return trimOuterQuotesFromArgs(result)
}

// Remove outer quotation marks from string
func trimOuterQuotesFromArgs(args []string) []string {
	var result []string
	for i := range args {
		arg := args[i]
		chars := []rune(arg)
		if unicode.In(chars[0], unicode.Quotation_Mark) && chars[0] == chars[len(chars)-1] {
			arg = arg[1 : len(arg)-1]
		}
		result = append(result, arg)
	}

	return result
}
